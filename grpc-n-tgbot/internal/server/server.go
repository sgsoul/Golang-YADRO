package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/core"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

//go:generate mockgen -source=server.go -destination=mocks/mock.go

type Service interface {
	Decode(w http.ResponseWriter, r *http.Request, v any)
	UpdateDatabase(workers int) (core.ComicCount, error)
	GetRateLimiter(ip string, rps int) *rate.Limiter
	CreateUserService(username, password, role string) error
	PrettyPrintService(comics []core.Comic) bytes.Buffer
	LimitedHandlerService(handler http.HandlerFunc) http.HandlerFunc
	GetUserByUsernameService(username string) (core.User, error)
}

type Search interface {
	RelevantURLS(str string, indexFile string) ([]string, []core.Comic)
}

type Server struct {
	config *core.Config
	service Service
	search Search
	authClient *AuthClient
}

func NewServer(cfg *core.Config, src Service, sr Search, authClient *AuthClient) (*Server, error) { // Update the function signature
	return &Server{
		search:     sr,
		config:     cfg,
		service:    src,
		authClient: authClient,
	}, nil
}

func (s *Server) Start() error {
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/register", s.handleRegister)
	http.HandleFunc("/pics", s.limitedHandler(s.rateLimitedHandler(s.handlePics)))
	http.HandleFunc("/update", s.limitedHandler(s.rateLimitedHandler(s.handleUpdate)))

	port := fmt.Sprintf(":%d", s.config.Port)
	fmt.Printf("\nServer listening on port %d\n", s.config.Port)
	return http.ListenAndServe(port, nil)
}

func (s *Server) handlePics(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getPics(w, r)
	default:
		http.Error(w, "invalid http method", http.StatusMethodNotAllowed)
	}
}

func (s *Server) getPics(w http.ResponseWriter, r *http.Request) {
	searchString := r.URL.Query().Get("search")
	if searchString == "" {
		http.Error(w, "search parameter is required", http.StatusBadRequest)
		return
	}

	comicURLs, _ := s.search.RelevantURLS(searchString, s.config.IndexFile)

	// JSON output
	urlJSON, err := json.Marshal(comicURLs)
	if err != nil {
		http.Error(w, "error marshaling comic URLs to JSON", http.StatusInternalServerError)
		return
	}

	fmt.Println("Response JSON:", string(urlJSON))

	w.Header().Set("Content-Type", "application/json")
	_ , err = w.Write(urlJSON)
	if err != nil{
		http.Error(w, err.Error(), http.StatusTeapot)
		return
	}

	// pretty output
	// responseBuffer := s.service.PrettyPrintService(relevantComics)

	// w.Header().Set("Content-Type", "text/plain")
	// _, err = w.Write(responseBuffer.Bytes())
	// if err != nil{
	// 	http.Error(w, err.Error(), http.StatusTeapot)
	// 	return
	// }
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:

		if !s.authClient.IsAdmin(w, r) {
			http.Error(w, "forbidden. administration rights required", http.StatusForbidden)
			return
		}

		response, err := s.service.UpdateDatabase(s.config.Parallel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(response)
		if err != nil{
			http.Error(w, err.Error(), http.StatusTeapot)
			return
		}
	default:
		http.Error(w, "invalid http method", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	user, err := s.service.GetUserByUsernameService(credentials.Username)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := s.authClient.GenerateJWT(user.Username, user.Role, s.config.TokenTime)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: time.Now().Add(time.Duration(s.config.TokenTime) * time.Minute),
	})
	w.WriteHeader(http.StatusOK)
}

func (s *Server) limitedHandler(handler http.HandlerFunc) http.HandlerFunc {
	return s.service.LimitedHandlerService(handler)
}

func (s *Server) rateLimitedHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := s.service.GetRateLimiter(ip, s.config.RateLim)
		if !limiter.Allow() {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		handler(w, r)
	}
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "invalid http method", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return 
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}

	if req.Role == "admin" {
		if !s.authClient.IsAdmin(w, r) {
			http.Error(w, "forbidden. administration rights required", http.StatusForbidden)
			return
		}
	}

	err = s.service.CreateUserService(req.Username, string(hashedPassword), req.Role)
	if err != nil {
		http.Error(w, "error saving user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func StartServer(cfg *core.Config, src Service, srch Search, authClient *AuthClient) { 
	xkcdServer, err := NewServer(cfg, src, srch, authClient)
	if err != nil {
		log.Error().Err(err).Msg("Error creating server")
	}

	if err := xkcdServer.Start(); err != nil {
		log.Error().Err(err).Msg("Error starting server")
	}
}