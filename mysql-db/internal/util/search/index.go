package search

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
    
	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/core/database"
)

func RelevantComic(relevantComics map[int]int, db *database.DB) ([]database.Comic, error) {
	var sortedComics []database.Comic

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
		comic, err := db.GetComicByID(item.Key)
		if err != nil {
			log.Printf("Error getting comic with ID %d: %v\n", item.Key, err)
			continue
		}
		sortedComics = append(sortedComics, comic)
	}

	return sortedComics, nil
}

func IndexSearch(file []byte, normalizedKeywords []string) (map[int]int) {
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

func BuildIndex(db *database.DB, indexFile string) error {
    log.Info().Msg("Building index...")

    // Получаем комиксы из базы данных
	comics, err := db.GetAllComics()
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

    log.Info().Msg("Index built successfully.")

	return nil
}


func New(db *database.DB, indexFile string) []byte {
	BuildIndex(db, "index.json")
	index, err := os.ReadFile(indexFile)
	if err != nil {
		log.Error().Err(err).Msg("error loading index fail")
		return index
	}
    return index
}

