package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateJWT(t *testing.T) {
	username := "testuser"
	role := "admin"
	expiryMinutes := 15

	tokenString, err := GenerateJWT(username, role, expiryMinutes)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
}

func TestValidateJWT(t *testing.T) {
	username := "testuser"
	role := "admin"
	expiryMinutes := 15

	tokenString, err := GenerateJWT(username, role, expiryMinutes)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	claims, err := ValidateJWT(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, role, claims.Role)
}

func TestIsAdmin(t *testing.T) {
	// токен для администратора
	adminToken, err := GenerateJWT("admin_user", "admin", 15)
	if err != nil {
		t.Fatalf("failed to generate admin token: %v", err)
	}

	// токен для обычного пользователя
	userToken, err := GenerateJWT("regular_user", "user", 15)
	if err != nil {
		t.Fatalf("failed to generate user token: %v", err)
	}
	
	// HTTP запрос с токеном администратора
	adminReq := httptest.NewRequest("GET", "/", nil)
	adminReq.AddCookie(&http.Cookie{Name: "token", Value: adminToken})
	adminRecorder := httptest.NewRecorder()

	if !IsAdmin(adminRecorder, adminReq) {
		t.Error("expected IsAdmin to return true for admin user")
	}

	// HTTP запрос с токеном обычного пользователя
	userReq := httptest.NewRequest("GET", "/", nil)
	userReq.AddCookie(&http.Cookie{Name: "token", Value: userToken})
	userRecorder := httptest.NewRecorder()

	if IsAdmin(userRecorder, userReq) {
		t.Error("expected IsAdmin to return false for non-admin user")
	}

	// HTTP запрос без токена
	reqWithoutToken := httptest.NewRequest("GET", "/", nil)
	reqWithoutTokenRecorder := httptest.NewRecorder()

	if IsAdmin(reqWithoutTokenRecorder, reqWithoutToken) {
		t.Error("expected IsAdmin to return false when token is missing")
	}

	// HTTP запрос с неправильным токеном
	invalidReq := httptest.NewRequest("GET", "/", nil)
	invalidReq.AddCookie(&http.Cookie{Name: "token", Value: "invalidToken"})
	invalidRecorder := httptest.NewRecorder()

	if IsAdmin(invalidRecorder, invalidReq) {
		t.Error("expected IsAdmin to return false for invalid token")
		
	}
}