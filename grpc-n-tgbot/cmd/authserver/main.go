package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc"
	pb "github.com/sgsoul/internal/proto"
)

var jwtKey = []byte("my_secret_key")

type server struct {
	pb.UnimplementedAuthServiceServer
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

func (s *server) GenerateJWT(ctx context.Context, req *pb.GenerateJWTRequest) (*pb.GenerateJWTResponse, error) {
	expirationTime := time.Now().Add(time.Duration(req.ExpiryMinutes) * time.Minute)
	claims := &Claims{
		Username: req.Username,
		Role:     req.Role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, err
	}
	return &pb.GenerateJWTResponse{Token: tokenString}, nil
}

func (s *server) ValidateJWT(ctx context.Context, req *pb.ValidateJWTRequest) (*pb.ValidateJWTResponse, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(req.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return &pb.ValidateJWTResponse{Valid: false}, nil
	}
	return &pb.ValidateJWTResponse{
		Valid:    true,
		Username: claims.Username,
		Role:     claims.Role,
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &server{})
	fmt.Println("Auth Service is running on port :50051")
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v", err)
	}
}
