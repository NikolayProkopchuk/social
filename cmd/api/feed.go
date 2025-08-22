package main

import (
	"net/http"

	"github.com/NikolayProkopchuk/social/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Fetches the user feed
//	@Description	Fetches the user feed
//	@Tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			since	query		string	false	"Since"
//	@Param			until	query		string	false	"Until"
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			tags	query		string	false	"Tags"
//	@Param			search	query		string	false	"Search"
//	@Success		200		{object}	[]store.PostWithMetadata
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
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
