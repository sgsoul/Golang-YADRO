package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jasonlvhit/gocron"
	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/core"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

//go:generate mockgen -source=service.go -destination=mocks/mock.go

type ClientXKCD interface {
	RunWorkers(workers int)
}

type Storage interface {
	GetCount() (int, error)
	CreateUser(username, password, role string) error
	PrettyPrint(v []core.Comic) bytes.Buffer
	GetAllComics() ([]core.Comic, error)
	GetComicByID(id int) (core.Comic, error)
	GetUserByUsername(username string) (core.User, error)
	SaveComicToDatabase(comic core.Comic) error
}

type service struct {
	storage Storage
	client  ClientXKCD
}

func NewService(cfg *core.Config, st Storage, cl ClientXKCD) *service { //??
	initConcurrencyLimiter(int64(cfg.ConcLim))

	return &service{
		client:  cl,
		storage: st,
	}
}

var mu sync.Mutex
var Sem *semaphore.Weighted
var rateLimiters = make(map[string]*rate.Limiter)

func initConcurrencyLimiter(limit int64) {
	Sem = semaphore.NewWeighted(limit)
}

func (s *service) GetRateLimiter(ip string, rps int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	if limiter, exists := rateLimiters[ip]; exists {
		return limiter
	}
	limiter := rate.NewLimiter(rate.Limit(rps), rps)
	rateLimiters[ip] = limiter
	return limiter
}

func StartScheduler(hour int, min int, port int) {
	t := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hour, min, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	err := gocron.Every(1).Hour().From(&t).Do(updateServer, port)
	if err != nil {
		return
	}

	<-gocron.Start()
}

func updateServer(port int) {
	log.Info().Msg("Scheduled database update")

	url := fmt.Sprintf("http://localhost:%d/update", port)
	resp, err := http.Post(url, "", nil)
	if err != nil {
		log.Error().Err(err).Msg("Error executing request")
		return
	}
	resp.Body.Close()
}

func (s *service) Decode(w http.ResponseWriter, r *http.Request, v any) {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
}

func (s *service) UpdateDatabase(workers int) (core.ComicCount, error) {
	loadedComicsCountBefore, _ := s.storage.GetCount()

	s.client.RunWorkers(workers)

	loadedComicsCountAfter, _ := s.storage.GetCount()

	updatedComicsCount := loadedComicsCountAfter - loadedComicsCountBefore

	response := core.ComicCount{
		UpdatedComics: updatedComicsCount,
		TotalComics:   loadedComicsCountAfter,
	}

	return response, nil
}

func (s *service) PrettyPrintService(comics []core.Comic) bytes.Buffer {
	return s.storage.PrettyPrint(comics)
}

func (s *service) GetUserByUsernameService(username string) (core.User, error) {
	return s.storage.GetUserByUsername(username)
}

func (s *service) CreateUserService(username, password, role string) error {
	return s.storage.CreateUser(username, password, role)
}

func (s *service) LimitedHandlerService(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !Sem.TryAcquire(1) {
			http.Error(w, "too many requests", http.StatusTooManyRequests)
			return
		}
		defer Sem.Release(1)
		handler(w, r)
	}
}
