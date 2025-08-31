package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/NikolayProkopchuk/social/internal/mailer"
	"github.com/NikolayProkopchuk/social/internal/store"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// registerUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validator.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
	}

	if err := user.Password.Set(payload.Password); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	plainInviteCode := uuid.New().String()
	app.logger.Infof("Generated invite code for %s: %s", user.Email, plainInviteCode)
	hash := sha256.Sum256([]byte(plainInviteCode))
	inviteCode := hex.EncodeToString(hash[:])

	if err := app.store.Users.CreateAndInvite(r.Context(), user, inviteCode, app.config.mail.exp); err != nil {
		if errors.Is(err, store.ErrConflict) {
			app.conflictError(w, r, err)
			return
		}
		app.internalServerError(w, r, err)
		return
	}
	isProdEnv := app.config.env == "prodaction"
	vars := struct {
		username   string
		activationUrl string
	}{
		username:  user.Username,
		activationUrl: app.config.frontednURL + "/activate?code=" + plainInviteCode,
	}

	if err := app.mailer.Send(mailer.UserInviteTemplate, user.Username, user.Email, vars, !isProdEnv); err != nil {
		app.logger.Errorw("unable to send email", "error", err)
		app.internalServerError(w, r, err)
		return
	}
	
	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	Email    string `json:"email" validate:"required,email"`
}

// createTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		200		{string}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]
func (app *application) createTokenHandler(w http.ResponseWriter, r *http.Request) {
	var payload CreateUserTokenPayload
	if err := readJSON(w, r, &payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validator.Struct(payload); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.unauthorizedError(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
			}
			return
	}
	log.Printf("User found: %+v", user)
	if err := user.Password.Compare(payload.Password); err != nil {
		app.unauthorizedError(w, r, err)
		return
	}
	
	claims := jwt.MapClaims{
		"sub": user.ID,
		"aud": app.config.auth.tokenCfg.audience,
		"iss": app.config.auth.tokenCfg.issuer,
		"exp": time.Now().Add(app.config.auth.tokenCfg.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, token); err != nil {
		app.internalServerError(w, r, err)
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}
