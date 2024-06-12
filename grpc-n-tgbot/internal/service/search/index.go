package search

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/core"
)

func (s *search) RelevantComic(relevantComics map[int]int) ([]core.Comic, error) {
	var sortedComics []core.Comic

	// слайс для сортировки
	type kv struct {
		Key   int
		Value int
	}

	var sortedSlice []kv
	for k, v := range relevantComics {
		sortedSlice = append(sortedSlice, kv{k, v})
	}
	sort.Slice(sortedSlice, func(i, j int) bool {
		return sortedSlice[i].Value > sortedSlice[j].Value
	})

	for _, item := range sortedSlice {
		comic, err := s.storage.GetComicByID(item.Key)
		if err != nil {
			log.Printf("Error getting comic with ID %d: %v\n", item.Key, err)
			continue
		}
		sortedComics = append(sortedComics, comic)
	}

	return sortedComics, nil
}

func IndexSearch(file []byte, normalizedKeywords []string) map[int]int {

	var index map[string][]int
	err := json.Unmarshal(file, &index)
	if err != nil {
		log.Error().Err(err).Msg("error loading index fail")
		return nil
	}

	relevantComics := make(map[int]int)
	for _, keyword := range normalizedKeywords {
		if comics, ok := index[keyword]; ok {
			for _, comic := range comics {
				relevantComics[comic]++
			}
		}
	}

	return relevantComics
}

func (s *search) buildIndex(indexFile string) error {
	log.Info().Msg("Building index...")

	// Получаем комиксы из базы данных
	comics, err := s.storage.GetAllComics()
	if err != nil {
		log.Error().Err(err).Msg("error getting comics from database")
		return err
	}

	// Создаем индекс
	index := make(map[string][]int)
	for _, comic := range comics {
		keywords := strings.Split(comic.Keywords, ",")
		for _, keyword := range keywords {
			index[keyword] = append(index[keyword], comic.ID)
		}
	}

	// Создаем файл индекса
	file, err := os.Create(indexFile)
	if err != nil {
		log.Error().Err(err).Msg("error creating index file")
		return err
	}
	defer file.Close()

	// Кодируем данные в формат JSON и записываем их в файл
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(index); err != nil {
		log.Error().Err(err).Msg("error encoding index to JSON")
		return err
	}

	log.Info().Msgf("Index built successfully. File location: %s", indexFile)

	return nil
}

func (s *search) newIndex(indexFile string) ([]byte, error) {
	err := s.buildIndex(indexFile)
	if err != nil {
		log.Error().Err(err).Msg("error building index")
		return nil, err
	}
	index, err := os.ReadFile(indexFile)
	if err != nil {
		log.Error().Err(err).Msg("error loading index fail")
		return nil, err
	}
	return index, nil
}


