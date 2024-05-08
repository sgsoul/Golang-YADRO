package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jasonlvhit/gocron"
	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/config"
	"github.com/sgsoul/internal/core/database"
	"github.com/sgsoul/internal/adapters/xkcd"
	"github.com/sgsoul/internal/util/words"
	sr "github.com/sgsoul/internal/util/search"
)


type Server struct {
	config *config.Config
	db     map[string]database.Comic 
	client *xkcd.Client
}

func NewServer(cfg *config.Config) (*Server, error) {
	client := xkcd.NewClient(cfg.SourceURL)
	db := database.New(cfg.DBFile)

	return &Server{
		client: client,
		config: cfg,
		db:     db,
	}, nil
}

func (s *Server) Start() error {
	http.HandleFunc("/update", s.handleUpdate)
	http.HandleFunc("/pics", s.handlePics)

	port := fmt.Sprintf(":%d", s.config.Port)
	fmt.Printf("Server listening on port %d\n", s.config.Port)
	return http.ListenAndServe(port, nil)
}

//========================== get pics ===========================================================


func (s *Server) handlePics(w http.ResponseWriter, r *http.Request) {
	switch r.Method{
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
	
	normalizedKeywords := words.NormalizeWords(searchString)

	index := sr.New(s.db, s.config.IndexFile)

	relevantComics := sr.RelevantComic(sr.IndexSearch(index, normalizedKeywords), s.config.DBFile)

	var comicURLs []string
	for i, comic := range relevantComics {
		if i >= 10 {
			break
		}
		comicURLs = append(comicURLs, comic.URL)
	}

	// json вывод
	urlJSON, err := json.Marshal(comicURLs)
	if err != nil {
		http.Error(w, "error marshaling comic URLs to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(urlJSON)
	
	// красивый приятный глазу вывод
	var responseBuffer bytes.Buffer
	for i, comic := range relevantComics {
		if i >= 10 {
			break
		}
		responseBuffer.WriteString(fmt.Sprintf("\nComic %d: %s\n", i+1, comic.URL))
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write(responseBuffer.Bytes())
}

//========================== post update ========================================================


func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
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

func (s *Server) updateDatabase() (interface{}, error) {
	loadedComicsCountBefore := database.GetCount(s.config.DBFile)

	s.client.RunWorkers(s.config.Parallel, s.config.DBFile)

	loadedComicsCountAfter:= database.GetCount(s.config.DBFile)

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

//=================================================================================

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

	log.Info().Msg("Sheduled database update")
	cfg := config.New("config.yaml")

	url := fmt.Sprintf("http://localhost:%d/update", cfg.Port)
	resp, err := http.Post(url, "", nil)
	if err != nil {
		log.Error().Err(err).Msg("Error executing request")
		return
	}
	resp.Body.Close()
}
