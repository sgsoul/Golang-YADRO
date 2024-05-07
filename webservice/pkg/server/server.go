package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sgsoul/pkg/config"
	"github.com/sgsoul/pkg/database"
	sr "github.com/sgsoul/pkg/search"
	"github.com/sgsoul/pkg/words"
	"github.com/sgsoul/pkg/xkcd"
)


type Server struct {
	config *config.Config
	db     map[string]database.Comic 
	client *xkcd.Client
}

func NewServer(cfg *config.Config) (*Server, error) {
	client := xkcd.NewClient(cfg.SourceURL)
	db, _, err := database.LoadComicsFromFile(cfg.DBFile)
	if err != nil {
		return nil, fmt.Errorf("error loading comics from database file: %v", http.StatusInternalServerError)
	}

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

	sr.BuildIndex(s.db, "index.json")
	index, err := os.ReadFile(s.config.IndexFile)
	if err != nil {
		http.Error(w, "error loading index file", http.StatusInternalServerError)
		return
	}

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

//=========================================================================================


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
	// количество загруженных комиксов до обновления
	_, loadedComicsCountBefore, err := database.LoadComicsFromFile(s.config.DBFile)
	if err != nil {
		return nil, err
	}

	s.client.RunWorkers(s.config.Parallel, s.config.DBFile)

	// количество загруженных комиксов после обновления
	_, loadedComicsCountAfter, err := database.LoadComicsFromFile(s.config.DBFile)
	if err != nil {
		return nil, err
	}

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