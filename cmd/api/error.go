package main

import (
	"log"
	"net/http"
)

func (app *application) internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("internal server error in request: %s %s error: %s\n", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequestError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("bad request: %s %s error: %s\n", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusBadRequest, err.Error())
}

func (app *application) conflictError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("conflict: %s %s error: %s\n", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusConflict, err.Error())
}

func (app *application) resourceNotFound(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("resourse not found: %s %s error: %s\n", r.Method, r.URL.Path, err)
	writeJSONError(w, http.StatusNotFound, "resource not found")
}
