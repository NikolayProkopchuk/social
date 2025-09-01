package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
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

type contextKey string

const userContextKey contextKey = "user"

func (app *application) authTokentMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthorizedError(w, r, fmt.Errorf("authorization header is required"))
			return
		}
		headerParts := strings.Split(authHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.unauthorizedError(w, r, fmt.Errorf("invalid authorization header format"))
			return
		}
		tokenString := headerParts[1]
		token, err := app.authenticator.ValidateToken(tokenString)
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("invalid token: %v", err))
			return
		}
		claims, _ := token.Claims.(jwt.MapClaims)
		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("invalid sub claim value"))
			return
		}
		ctx := r.Context()
		user, err := app.store.Users.GetByID(ctx, userID)
		if err != nil {
			app.unauthorizedError(w, r, fmt.Errorf("user not found"))
			return
		}
		ctx = context.WithValue(ctx, userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
