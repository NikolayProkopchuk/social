package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/NikolayProkopchuk/social/internal/store"
	"github.com/go-chi/chi/v5"
)

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserFromContext(r)
	if err := app.jsonResponse(w, http.StatusOK, user); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) userContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}
		user, err := app.store.Users.GetByID(r.Context(), userID)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.resourceNotFound(w, r, err)
				return
			default:
				app.internalServerError(w, r, err)
			}
			return
		}
		ctx := context.WithValue(r.Context(), "user", user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) getUserFromContext(r *http.Request) *store.User {
	value := r.Context().Value("user")
	return value.(*store.User)
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserFromContext(r)
	//todo get authenticated user
	userLoggedIn := &store.User{
		ID: 2,
	}

	if err := app.store.Followers.Follow(r.Context(), user, userLoggedIn); err != nil {
		switch {
		case errors.Is(err, store.ErrConflict):
			app.conflictError(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserFromContext(r)
	//todo get authenticated user
	userLoggedIn := &store.User{
		ID: 2,
	}

	if err := app.store.Followers.Unfollow(r.Context(), user, userLoggedIn); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusNoContent, nil); err != nil {
		app.internalServerError(w, r, err)
	}
}
