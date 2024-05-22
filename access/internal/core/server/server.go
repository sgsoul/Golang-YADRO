package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/sgsoul/internal/adapters/auth"
	"github.com/sgsoul/internal/adapters/xkcd"
	"github.com/sgsoul/internal/config"
	"github.com/sgsoul/internal/core/database"
	"golang.org/x/crypto/bcrypt"
)

type Server struct {
	config *config.Config
	db     *database.DB
	client *xkcd.Client
}

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewServer(cfg *config.Config) (*Server, error) {
	client := xkcd.NewClient(cfg.SourceURL)
	db := database.NewDB(cfg.DSN)

	initConcurrencyLimiter(int64(cfg.ConcLim))

	return &Server{
		client: client,
		config: cfg,
		db:     db,
	}, nil
}

func (s *Server) Start() error {
	http.HandleFunc("/update", s.limitedHandler(s.rateLimitedHandler(s.handleUpdate)))
	http.HandleFunc("/pics", s.limitedHandler(s.rateLimitedHandler(s.handlePics)))
	http.HandleFunc("/login", s.handleLogin)
	http.HandleFunc("/register", s.handleRegister)

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

	comicURLs, relevantComics := relevantURLS(searchString, s.config.IndexFile, s.db)

	// JSON output
	urlJSON, err := json.Marshal(comicURLs)
	if err != nil {
		http.Error(w, "error marshaling comic URLs to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(urlJSON)

	// pretty output
	responseBuffer := database.PrettyPrint(relevantComics)

	w.Header().Set("Content-Type", "text/plain")
	w.Write(responseBuffer.Bytes())
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "no token provided", http.StatusUnauthorized)
			return
		}
	
		claims, err := auth.ValidateJWT(tokenCookie.Value)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
	
		if claims.Role != "admin" {
			http.Error(w, "forbidden. administration rights required", http.StatusForbidden)
			return
		}

		response, err := s.updateDatabase()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	default:
		http.Error(w, "invalid http method", http.StatusMethodNotAllowed)
	}
}

// сервис?? целиком? [4]
func (s *Server) updateDatabase() (interface{}, error) {
	loadedComicsCountBefore, _ := s.db.GetCount()

	s.client.RunWorkers(s.config.Parallel, s.db)

	loadedComicsCountAfter, _ := s.db.GetCount()

	updatedComicsCount := loadedComicsCountAfter - loadedComicsCountBefore

	response := struct {
		UpdatedComics int `json:"updated_comics"`
		TotalComics   int `json:"total_comics"`
	}{
		UpdatedComics: updatedComicsCount,
		TotalComics:   loadedComicsCountAfter,
	}

	return response, nil
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var credentials struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	Decode(w, r, &credentials)

	user, err := s.db.GetUserByUsername(credentials.Username)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// это в сервис? [1]
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateJWT(user.Username, user.Role, s.config.TokenTime)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// это в сервис? [3]
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   token,
		Expires: time.Now().Add(time.Duration(s.config.TokenTime) * time.Minute),
	})
	w.WriteHeader(http.StatusOK)
}

func (s *Server) limitedHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !sem.TryAcquire(1) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		defer sem.Release(1)
		handler(w, r)
	}
}

func (s *Server) rateLimitedHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		limiter := getRateLimiter(ip, s.config.RateLim)
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
	Decode(w, r, &req)

	// это в сервис? [2]
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error hashing password", http.StatusInternalServerError)
		return
	}

	if req.Role == "admin" {
		tokenCookie, err := r.Cookie("token")
		if err != nil {
			http.Error(w, "no token provided", http.StatusUnauthorized)
			return
		}
	
		claims, err := auth.ValidateJWT(tokenCookie.Value)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
	
		if claims.Role != "admin" {
			http.Error(w, "forbidden. administration rights required", http.StatusForbidden)
			return
		}
	}

	err = s.db.CreateUser(req.Username, string(hashedPassword), req.Role)
	if err != nil {
		http.Error(w, "error saving user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
