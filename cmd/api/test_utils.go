package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NikolayProkopchuk/social/internal/auth"
	"github.com/NikolayProkopchuk/social/internal/store"
	"github.com/NikolayProkopchuk/social/internal/store/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func newTestApp(t *testing.T, cacheEnabled bool) *application {
	t.Helper()

	logger := zap.Must(zap.NewDevelopment()).Sugar()

	moderatorUser := store.User{
		ID:       1,
		Username: "TestModeratorUser",
		Email:    "test.moderator@mail.com",
		Role:     store.Role{ID: 2, Name: "moderator", Description: "Moderator", Level: 50},
	}
	user1 := store.User{
		ID:       2,
		Username: "TestUser",
		Email:    "test@mail.com",
		Role:     store.Role{ID: 3, Name: "user", Description: "Regular User", Level: 10},
	}
	user3 := store.User{
		ID:       3,
		Username: "TestUser3",
		Email:    "test3@mail.com",
		Role:     store.Role{ID: 3, Name: "user", Description: "Regular User", Level: 10},
	}

	mockCache := cache.NewMockCache()
	mockUserCache := mockCache.Users.(*cache.MockUserCache)
	mockUserCache.On("Get", mock.Anything, int64(1)).Return(&moderatorUser, nil)
	mockUserCache.On("Get", mock.Anything, int64(2)).Return(&user1, nil)
	mockUserCache.On("Get", mock.Anything, int64(3)).Return(nil, store.ErrNotFound)

	mockUserCache.On("Set", mock.Anything, user3).Return(nil)

	mockStore := store.NewMockStore()
	mockUserStore := mockStore.Users.(*store.MockUserStore)
	mockUserStore.On("GetByID", mock.Anything, int64(1)).Return(&moderatorUser, nil)
	mockUserStore.On("GetByID", mock.Anything, int64(2)).Return(&user1, nil)
	mockUserStore.On("GetByID", mock.Anything, int64(3)).Return(&user3, nil)

	mockRoleStore := mockStore.Roles.(*store.MockRoleStore)
	mockRoleStore.On("GetByName", mock.Anything, "moderator").Return(
		&store.Role{ID: 2, Name: "moderator", Description: "Moderator", Level: 50}, nil)

	app := &application{
		logger:        logger,
		store:         mockStore,
		cache:         mockCache,
		authenticator: auth.NewMockAuthenticator(),
		config: config{
			redis: &redisConfig{
				enabled: cacheEnabled,
			},
		},
	}

	return app
}

func executeRequest(req *http.Request, mux http.Handler) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	t.Helper()
	assert.Equal(t, expected, actual)
}

func checkResponseBody(t *testing.T, expected map[string]any, actual []byte) {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal(actual, &got); err != nil {
		t.Fatalf("invalid json response: %v; body: %s", err, string(actual))
	}
	assert.Equal(t, expected, got)
}
