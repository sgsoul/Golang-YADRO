package search

import (
	"encoding/json"
	"os"
	"sort"
	"strconv"

	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/pkg/database"
)

func RelevantComic(relevantComics map[int]int, dbFile string) []database.Comic {
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

    // комиксы в порядке убывания количества совпадающих ключевых слов
    for _, item := range sortedSlice {
        comicList, err := database.GetComic(dbFile, item.Key)
        if err != nil {
            log.Error().Err(err).Msg("error during getting comic")
            return nil
        }
        sortedComics = append(sortedComics, comicList...)
    }

    return sortedComics
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

func BuildIndex(db map[string]database.Comic, indexFile string) error {
	index := make(map[string][]int)

	for id, comic := range db {
		for _, keyword := range comic.Keywords {
			idInt, _ := strconv.Atoi(id)
			index[keyword] = append(index[keyword], idInt)
		}
	}

	file, err := os.Create(indexFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// JSON-кодировщик для записи данных в файл
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ") // отступ для читаемости

	if err := encoder.Encode(index); err != nil {
		return err
	}

	return nil
}
