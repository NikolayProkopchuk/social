package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/NikolayProkopchuk/social/internal/store"
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
	var payload RegistrerUserPayload
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
	if err := app.jsonResponse(w, http.StatusCreated, user); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

type RegistrerUserPayload struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	Email    string `json:"email" validate:"required,email"`
}
