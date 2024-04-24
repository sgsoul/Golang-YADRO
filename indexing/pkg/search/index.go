package search

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/sgsoul/pkg/database"
)

func RelevantComic(ids []int, dbFile string) []database.Comic {
	var relevantComics []database.Comic

	for _, id := range ids {
		comicList, err := database.GetComic(dbFile, id)
		if err != nil {
			fmt.Println("error during getting comic:", err)
			return nil
		}
		relevantComics = append(relevantComics, comicList...)
	}

	return relevantComics
}

func IndexSearch(file []byte, normalizedKeywords []string) []int {
	var index map[string][]int
	err := json.Unmarshal(file, &index)
	if err != nil {
		fmt.Println("error loading index fail:", err)
	}

	relevantComics := make(map[int]struct{})
	for _, keyword := range normalizedKeywords {
		if comics, ok := index[keyword]; ok {
			for _, comic := range comics {
				relevantComics[comic] = struct{}{}
			}
		}
	}

	result := make([]int, 0, len(relevantComics))
	for comic := range relevantComics {
		result = append(result, comic)
	}

	return result
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
