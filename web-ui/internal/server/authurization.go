package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
)

var jwtKey = []byte("my_secret_key")

type Claims struct { 
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func GenerateJWT(username, role string, expiryMinutes int) (string, error) {
	expirationTime := time.Now().Add(time.Duration(expiryMinutes) * time.Minute)
	claims := &Claims{
		Username: username,
		Role:     role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return claims, nil
}

func IsAdmin(w http.ResponseWriter, r *http.Request) bool {
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "no token provided", http.StatusUnauthorized)
		return false
	}

	claims, err := ValidateJWT(tokenCookie.Value)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return false
	}

	if claims.Role != "admin" {
		return false
	}

	return true
}
