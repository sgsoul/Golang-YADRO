package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/jasonlvhit/gocron"
	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/config"
	"github.com/sgsoul/internal/core/database"
	sr "github.com/sgsoul/internal/util/search"
	"github.com/sgsoul/internal/util/words"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
)

var sem *semaphore.Weighted

func initConcurrencyLimiter(limit int64) {
	sem = semaphore.NewWeighted(limit)
}

var rateLimiters = make(map[string]*rate.Limiter)
var mu sync.Mutex

func getRateLimiter(ip string, rps int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()
	if limiter, exists := rateLimiters[ip]; exists {
		return limiter
	}
	limiter := rate.NewLimiter(rate.Limit(rps), rps)
	rateLimiters[ip] = limiter
	return limiter
}

func StartScheduler(hour int, min int) {
	t := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), hour, min, 0, 0, time.FixedZone("UTC+3", 3*60*60))
	gocron.Every(1).Hour().From(&t).Do(update)

	<-gocron.Start()
}

func StartServer(cfg *config.Config) {
	xkcdServer, err := NewServer(cfg)
	if err != nil {
		log.Error().Err(err).Msg("Error creating server")
	}

	if err := xkcdServer.Start(); err != nil {
		log.Error().Err(err).Msg("Error starting server")
	}
}

func update() {
	log.Info().Msg("Scheduled database update")
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Error().Msg("error")
		return
	}

	url := fmt.Sprintf("http://localhost:%d/update", cfg.Port)
	resp, err := http.Post(url, "", nil)
	if err != nil {
		log.Error().Err(err).Msg("Error executing request")
		return
	}
	resp.Body.Close()
}

func Decode(w http.ResponseWriter, r *http.Request, v any) {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
}

func relevantURLS(str string, indexFile string, db *database.DB) ([]string, []database.Comic) {
	normalizedKeywords := words.NormalizeWords(str)

	index := sr.New(db, indexFile)

	relevantComics, err := sr.RelevantComic(sr.IndexSearch(index, normalizedKeywords), db)
	if err != nil {
		log.Error().Msg("error ")
		return nil, nil
	}

	var comicURLs []string
	for i, comic := range relevantComics {
		if i >= 10 {
			break
		}
		comicURLs = append(comicURLs, comic.URL)
	}

	return comicURLs, relevantComics
}
