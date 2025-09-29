package main

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/NikolayProkopchuk/social/internal/ratelimiter"
	"github.com/NikolayProkopchuk/social/internal/store"
	"github.com/NikolayProkopchuk/social/internal/store/cache"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/mock"
)

var claims = jwt.MapClaims{
	"sub": 42,
	"aud": "test_aud",
	"iss": "test_iss",
	"exp": time.Now().Add(time.Hour).Unix(),
	"iat": time.Now().Unix(),
	"nbf": time.Now().Unix()}

var expectedModeratorResponseBody = map[string]any{
	"data": map[string]any{
		"id":         float64(1),
		"email":      "test.moderator@mail.com",
		"username":   "TestModeratorUser",
		"created_at": "0001-01-01T00:00:00Z",
		"role": map[string]any{
			"id":          float64(2),
			"name":        "moderator",
			"level":       float64(50),
			"description": "Moderator",
		},
	},
}

var expectedUser1ResponseBody = map[string]any{
	"data": map[string]any{
		"id":         float64(2),
		"email":      "test@mail.com",
		"username":   "TestUser",
		"created_at": "0001-01-01T00:00:00Z",
		"role": map[string]any{
			"id":          float64(3),
			"name":        "user",
			"level":       float64(10),
			"description": "Regular User",
		},
	},
}

var expectedUser3ResponseBody = map[string]any{
	"data": map[string]any{
		"id":         float64(3),
		"email":      "test3@mail.com",
		"username":   "TestUser3",
		"created_at": "0001-01-01T00:00:00Z",
		"role": map[string]any{
			"id":          float64(3),
			"name":        "user",
			"level":       float64(10),
			"description": "Regular User",
		},
	},
}

func TestGetUser(t *testing.T) {
	cfg := config{
		redis: &redisConfig{
			enabled: false,
		},
		rateLimiter: &ratelimiter.Config{
			Enabled: false,
		},
	}
	app := newTestApp(t, cfg)
	mux := app.mount()

	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		t.Fatal(err)
	}
	authHeader := fmt.Sprintf("Bearer %s", token)

	test := []struct {
		name                string
		userID              int64
		expectedStatus      int
		authorizationHeader string
		expectedBody        map[string]any
	}{
		{
			name:                "shoud not allow unauthenticated requests",
			userID:              2,
			expectedStatus:      http.StatusUnauthorized,
			authorizationHeader: "",
			expectedBody: map[string]any{
				"error": "authorization header is required",
			},
		},
		{
			name:                "shoud allow authenticated requests",
			userID:              2,
			expectedStatus:      http.StatusOK,
			authorizationHeader: authHeader,
			expectedBody:        expectedUser1ResponseBody,
		},
	}

	for _, tc := range test {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", fmt.Sprintf("/v1/users/%d", tc.userID), nil)
			if err != nil {
				t.Fatal(err)
			}

			if tc.authorizationHeader != "" {
				req.Header.Set("Authorization", tc.authorizationHeader)
			}

			rr := executeRequest(req, mux)
			checkResponseCode(t, tc.expectedStatus, rr.Code)
			if tc.expectedBody != nil {
				checkResponseBody(t, tc.expectedBody, rr.Body.Bytes())
			}
		})

	}
}

func TestGetUserStorageCalls(t *testing.T) {

	t.Run("should call cashe twice when it is enabled and no call the store", func(t *testing.T) {
		cfg := config{
			redis: &redisConfig{
				enabled: true,
			},
			rateLimiter: &ratelimiter.Config{
				Enabled: false,
			},
		}
		app := newTestApp(t, cfg)
		mux := app.mount()
		token, err := app.authenticator.GenerateToken(claims)
		if err != nil {
			t.Fatal(err)
		}
		authHeader := fmt.Sprintf("Bearer %s", token)

		req, err := http.NewRequest("GET", fmt.Sprintf("/v1/users/%d", 2), nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", authHeader)
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)
		checkResponseBody(t, expectedUser1ResponseBody, rr.Body.Bytes())

		mockUserCache := app.cache.Users.(*cache.MockUserCache)
		mockUserCache.AssertNumberOfCalls(t, "Get", 2)

		mockUserStore := app.store.Users.(*store.MockUserStore)
		mockUserStore.AssertNumberOfCalls(t, "GetByID", 0)
	})

	t.Run("should not call cache when it is not enabled and call the user store twice and does not call role store if user get himself", func(t *testing.T) {
		cfg := config{
			redis: &redisConfig{
				enabled: false,
			},
			rateLimiter: &ratelimiter.Config{
				Enabled: false,
			},
		}
		app := newTestApp(t, cfg)
		mux := app.mount()
		token, err := app.authenticator.GenerateToken(claims)
		if err != nil {
			t.Fatal(err)
		}
		authHeader := fmt.Sprintf("Bearer %s", token)

		req, err := http.NewRequest("GET", fmt.Sprintf("/v1/users/%d", 1), nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", authHeader)
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)
		checkResponseBody(t, expectedModeratorResponseBody, rr.Body.Bytes())

		mockUserCache := app.cache.Users.(*cache.MockUserCache)
		mockUserCache.AssertNumberOfCalls(t, "Get", 0)

		mockUserStore := app.store.Users.(*store.MockUserStore)
		mockUserStore.AssertNumberOfCalls(t, "GetByID", 2)

		mockRoleStore := app.store.Roles.(*store.MockRoleStore)
		mockRoleStore.AssertNumberOfCalls(t, "GetByName", 0)
	})

	t.Run("should not call cashe when it is not enabled and call the user store twice and role store once if user moderator and get another user", func(t *testing.T) {
		cfg := config{
			redis: &redisConfig{
				enabled: false,
			},
			rateLimiter: &ratelimiter.Config{
				Enabled: false,
			},
		}
		app := newTestApp(t, cfg)
		mux := app.mount()
		token, err := app.authenticator.GenerateToken(claims)
		if err != nil {
			t.Fatal(err)
		}
		authHeader := fmt.Sprintf("Bearer %s", token)

		req, err := http.NewRequest("GET", fmt.Sprintf("/v1/users/%d", 2), nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", authHeader)
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)
		checkResponseBody(t, expectedUser1ResponseBody, rr.Body.Bytes())

		mockUserCache := app.cache.Users.(*cache.MockUserCache)
		mockUserCache.AssertNumberOfCalls(t, "Get", 0)

		mockUserStore := app.store.Users.(*store.MockUserStore)
		mockUserStore.AssertNumberOfCalls(t, "GetByID", 2)

		mockRoleStore := app.store.Roles.(*store.MockRoleStore)
		mockRoleStore.AssertNumberOfCalls(t, "GetByName", 1)
	})

	t.Run("should put user in cache and return it", func(t *testing.T) {
		cfg := config{
			redis: &redisConfig{
				enabled: true,
			},
			rateLimiter: &ratelimiter.Config{
				Enabled: false,
			},
		}
		app := newTestApp(t, cfg)
		mux := app.mount()

		token, err := app.authenticator.GenerateToken(claims)
		if err != nil {
			t.Fatal(err)
		}
		authHeader := fmt.Sprintf("Bearer %s", token)

		req, err := http.NewRequest("GET", fmt.Sprintf("/v1/users/%d", 3), nil)
		if err != nil {
			t.Fatal(err)
		}

		req.Header.Set("Authorization", authHeader)
		rr := executeRequest(req, mux)

		checkResponseCode(t, http.StatusOK, rr.Code)
		checkResponseBody(t, expectedUser3ResponseBody, rr.Body.Bytes())

		mockUserCache := app.cache.Users.(*cache.MockUserCache)
		mockUserCache.AssertCalled(t, "Get", mock.Anything, int64(1))
		mockUserCache.AssertCalled(t, "Get", mock.Anything, int64(3))
		mockUserCache.AssertCalled(t, "Set", mock.Anything, mock.Anything)

		mockUserStore := app.store.Users.(*store.MockUserStore)
		mockUserStore.AssertCalled(t, "GetByID", mock.Anything, int64(3))

		mockRoleStore := app.store.Roles.(*store.MockRoleStore)
		mockRoleStore.AssertCalled(t, "GetByName", mock.Anything, "moderator")
	})
}
