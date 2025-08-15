package main

import (
	"net/http"

	"github.com/NikolayProkopchuk/social/internal/store"
)

func (app *application) getUserFeedHandler(w http.ResponseWriter, r *http.Request) {
	user := store.User{
		ID: 1,
	}
	paginatedFeedQuery, err := store.ParsePaginatedFeedQuery(r)
	if err != nil {
		app.badRequestError(w, r, err)
	}
	err = Validator.Struct(paginatedFeedQuery)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}
	feed, err := app.store.Posts.GetUserFeed(r.Context(), &user, paginatedFeedQuery)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	if err = app.jsonResponse(w, http.StatusOK, feed); err != nil {
		app.internalServerError(w, r, err)
	}
}
