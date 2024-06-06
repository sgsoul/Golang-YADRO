package search

import (
	"sort"
	"strings"

	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/core"
	"github.com/sgsoul/internal/words"
)

//go:generate mockgen -source=search.go -destination=mocks/mock.go

type Storage interface {
	GetAllComics() ([]core.Comic, error)
	GetComicByID(id int) (core.Comic, error)
}

type search struct {
	storage Storage
}

func NewSearch(st Storage) *search { //??
	return &search{
		storage: st,
	}
}

func (s *search) FindRelevantComics(keywords []string) ([]core.Comic, error) {
	var relevantComics []core.Comic

	comics, err := s.storage.GetAllComics()
	if err != nil {
		return nil, err
	}

	for _, comic := range comics {
		relevanceScore := countMatchingKeywords(comic, keywords)

		if relevanceScore > 0 {
			relevantComics = append(relevantComics, comic)
		}
	}

	sort.Slice(relevantComics, func(i, j int) bool {
		return countMatchingKeywords(relevantComics[i], keywords) > countMatchingKeywords(relevantComics[j], keywords)
	})

	return relevantComics, nil
}

func countMatchingKeywords(comic core.Comic, keywords []string) int {
	var count int
	for _, keyword := range keywords {
		normalizedKeywords := words.NormalizeWords(keyword)
		for _, normalizedKeyword := range normalizedKeywords {
			comicKeywords := strings.Split(comic.Keywords, ",")
			for _, comicKeyword := range comicKeywords {
				if strings.EqualFold(comicKeyword, normalizedKeyword) {
					count++
					break
				}
			}
		}
	}
	return count
}

func (s *search) RelevantURLS(str string, indexFile string) ([]string, []core.Comic) {
	normalizedKeywords := words.NormalizeWords(str)

	index, _ := s.newIndex(indexFile)

	relevantComics, err := s.RelevantComic(IndexSearch(index, normalizedKeywords))
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
