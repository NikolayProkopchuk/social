package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/NikolayProkopchuk/social/internal/store"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
)

func (app *application) basicAuthMiddleware() func(http.Handler) http.Handler {
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
		user, err := app.getUser(ctx, userID)
		if err != nil {
			app.logger.Warn(err)
			app.unauthorizedError(w, r, fmt.Errorf("user not found"))
			return
		}
		ctx = context.WithValue(ctx, userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) postOwnershipMiddleware(roleName string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserFromContext(r)
		post := app.getPostFromContext(r)
		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}
		role, err := app.store.Roles.GetByName(r.Context(), roleName)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if user.Role.Level >= role.Level {
			next.ServeHTTP(w, r)
			return
		}
		app.resourceForbiddenError(w, r, fmt.Errorf("post modification is allowed only for owner or users with %s role", roleName))
	})
}

func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {
	if !app.config.redis.enabled {
		return app.store.Users.GetByID(ctx, userID)
	}
	user, err := app.cache.Users.Get(ctx, userID)
	if err == nil {
		app.logger.Infow("User found in cache", "userID", userID)
		return user, nil
	}
	if err != redis.Nil {
		app.logger.Errorw("Failed to fetch user from cache", "userID", userID, "error", err)
	}
	app.logger.Infow("User not found in cache, fetching from DB", "userID", userID)
	user, err = app.store.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	app.logger.Infow("Caching user", "userID", userID)
	err = app.cache.Users.Set(ctx, *user)
	if err != nil {
		app.logger.Errorw("Failed to cache user", "userID", userID, "error", err)
	}

	return user, err
}

func (app *application) userOwnershipMiddleware(roleName string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserFromContext(r)
		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
		if err != nil {
			app.badRequestError(w, r, err)
			return
		}
		role, err := app.store.Roles.GetByName(r.Context(), roleName)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		if user.ID != userID && user.Role.Level < role.Level {
			app.resourceForbiddenError(w, r, fmt.Errorf("user resource access is allowed only for owner or users with %s role", roleName))
			return
		}
		next.ServeHTTP(w, r)
	})
}
