package main

import (
	"fmt"
	"net/http"
)

func (app *application) basicAuthMiddleware() func (http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != app.config.auth.basic.username || password != app.config.auth.basic.password {
			app.unauthorizedBasicError(w, r, fmt.Errorf("unauthorized"))
			return
		}
		next.ServeHTTP(w, r)
	})
	}
}
