package server

import (
	"context"
	"net/http"
	"time"

	pb "github.com/sgsoul/internal/proto"
	"google.golang.org/grpc"
)

type AuthClient struct {
	client pb.AuthServiceClient
}

func NewAuthClient(address string) (*AuthClient, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	client := pb.NewAuthServiceClient(conn)
	return &AuthClient{client: client}, nil
}

func (a *AuthClient) GenerateJWT(username, role string, expiryMinutes int) (string, error) {
	var expiryMinutesInt32 int32 = int32(expiryMinutes)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.GenerateJWTRequest{
		Username:     username,
		Role:         role,
		ExpiryMinutes: expiryMinutesInt32,
	}
	res, err := a.client.GenerateJWT(ctx, req)
	if err != nil {
		return "", err
	}
	return res.Token, nil
}

func (a *AuthClient) ValidateJWT(token string) (*pb.ValidateJWTResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	req := &pb.ValidateJWTRequest{Token: token}
	res, err := a.client.ValidateJWT(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (a *AuthClient) IsAdmin(w http.ResponseWriter, r *http.Request) bool {
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		http.Error(w, "no token provided", http.StatusUnauthorized)
		return false
	}

	claims, err := a.ValidateJWT(tokenCookie.Value)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return false
	}

	if claims.Role != "admin" {
		return false
	}

	return true
}

