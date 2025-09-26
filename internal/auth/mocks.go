package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const secret = "test_secret"

var testClaims = jwt.MapClaims{
	"aud": "test_aud",
	"iss": "test_aud",
	"sub": int64(1),
	"exp": time.Now().Add(1 * time.Hour).Unix(),
	"iat": time.Now().Unix(),
	"nbf": time.Now().Unix(),
}

func NewMockAuthenticator() *MockAuthenticator {
	return &MockAuthenticator{}
}

type MockAuthenticator struct {
}

func (m *MockAuthenticator) GenerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, testClaims)
	return token.SignedString([]byte(secret))
}

func (m *MockAuthenticator) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
}
