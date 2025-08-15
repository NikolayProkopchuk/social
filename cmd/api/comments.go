package main

import (
	"net/http"

	"github.com/NikolayProkopchuk/social/internal/store"
)

type CommentPayload struct {
	Content string `json:"content" validate:"required,max=1000"`
}

func (app *application) createCommentHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	userID := 1
	commentPayload := &CommentPayload{}
	if err := readJSON(w, r, commentPayload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if err := Validator.Struct(commentPayload); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	comment := &store.Comment{
		PostID: post.ID,
		User: &store.User{
			ID: int64(userID),
		},
		Content: commentPayload.Content,
	}

	if err := app.store.Comments.Create(r.Context(), comment); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, comment); err != nil {
		app.internalServerError(w, r, err)
	}
}
