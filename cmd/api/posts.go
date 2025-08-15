package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/NikolayProkopchuk/social/internal/store"
	"github.com/go-chi/chi/v5"
)

type postKey string

const postCtx postKey = "post"

type createPostRequest struct {
	Title   string   `json:"title" validate:"required,max=100"`
	Content string   `json:"content" validate:"required,max=10000"`
	Tags    []string `json:"tags"`
}

func (app *application) createPostHandler(w http.ResponseWriter, r *http.Request) {
	var createPostDto createPostRequest
	if err := readJSON(w, r, &createPostDto); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err := Validator.Struct(createPostDto); err != nil {
		app.badRequestError(w, r, err)
		return
	}

	post := store.Post{
		UserID:  2,
		Title:   createPostDto.Title,
		Content: createPostDto.Content,
		Tags:    createPostDto.Tags,
	}

	if err := app.store.Posts.Create(r.Context(), &post); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := app.jsonResponse(w, http.StatusCreated, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) getPostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)
	comments, err := app.store.Comments.GetByPostID(r.Context(), post.ID)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	post.Comments = comments
	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

type updatePostRequest struct {
	Title   *string  `json:"title" validate:"omitempty,max=100"`
	Content *string  `json:"content" validate:"omitempty,max=10000"`
	Tags    []string `json:"tags"`
}

func (app *application) updatePostHandler(w http.ResponseWriter, r *http.Request) {
	post := getPostFromCtx(r)

	var updatePostDto updatePostRequest

	if err := readJSON(w, r, &updatePostDto); err != nil {
		app.badRequestError(w, r, err)
		return
	}
	if updatePostDto.Title != nil {
		post.Title = *updatePostDto.Title
	}
	if updatePostDto.Content != nil {
		post.Content = *updatePostDto.Content
	}
	if updatePostDto.Tags != nil {
		post.Tags = updatePostDto.Tags
	}
	if err := app.store.Posts.Update(r.Context(), post); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.resourceNotFound(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	if err := app.jsonResponse(w, http.StatusOK, post); err != nil {
		app.internalServerError(w, r, err)
	}
}

func (app *application) deletePostHandler(w http.ResponseWriter, r *http.Request) {
	postIDParam := chi.URLParam(r, "postID")

	postID, err := strconv.ParseInt(postIDParam, 10, 64)
	if err != nil {
		app.badRequestError(w, r, err)
		return
	}

	if err = app.store.Posts.DeleteByID(r.Context(), postID); err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.resourceNotFound(w, r, err)
			return
		default:
			app.internalServerError(w, r, err)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (app *application) postsContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postIDParam := chi.URLParam(r, "postID")

		postID, err := strconv.ParseInt(postIDParam, 10, 64)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}

		ctx := r.Context()

		post, err := app.store.Posts.GetByID(ctx, postID)
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
		ctx = context.WithValue(ctx, postCtx, post)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func getPostFromCtx(r *http.Request) *store.Post {
	post, _ := r.Context().Value(postCtx).(*store.Post)
	return post
}
