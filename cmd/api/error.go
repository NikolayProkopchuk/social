package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal server error",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)
	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict: %s %s error: %s\n", r.Method, r.URL.Path, err)
	app.logger.Errorf("conflict",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)
	writeJSONError(w, http.StatusConflict, err.Error())
}

func (app *application) resourceNotFound(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("resourse not found: %s %s error: %s\n", r.Method, r.URL.Path, err)
	app.logger.Warnf("resource not found",
		"method", r.Method,
		"path", r.URL.Path,
		"error", err.Error(),
	)
	writeJSONError(w, http.StatusNotFound, "resource not found")
}
