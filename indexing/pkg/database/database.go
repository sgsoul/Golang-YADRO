package database

import (
	"encoding/json"
	"fmt"
	"os"
)

type Comic struct {
	URL      string   `json:"url"`
	Keywords []string `json:"keywords"`
}

func LoadComicsFromFile(dbFile string) (map[string]Comic, error) {
	file, err := os.Open(dbFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var comicsMap = make(map[string]Comic)

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var comicData map[string]Comic
		if err := decoder.Decode(&comicData); err != nil {
			return nil, err
		}
		for key, comic := range comicData {
			comicsMap[key] = comic
		}
	}

	//fmt.Println("DEBUG num of loaded comics:", len(comicsMap))

	return comicsMap, nil
}

func SaveComicsToDatabase(dbFile string, comic Comic, i int) error {
	// открываем дб файл
	file, err := os.OpenFile(dbFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// сериализуем и добавляем в дб
	comicJSON, err := json.MarshalIndent(map[string]Comic{fmt.Sprintf("%d", i): comic}, "", "   ")
	if err != nil {
		return err
	}
	_, err = file.Write(comicJSON)
	if err != nil {
		return err
	}
	_, err = file.WriteString("\n")
	if err != nil {
		return err
	}

	return nil
}

func GetComic(dbFile string, id int) ([]Comic, error) {
	file, err := os.Open(dbFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var comicData map[string]Comic
		if err := decoder.Decode(&comicData); err != nil {
			return nil, err
		}
		comic, ok := comicData[fmt.Sprintf("%d", id)]
		if ok {
			return []Comic{comic}, nil
		}
	}

	return nil, fmt.Errorf("comic with ID %d not found", id)
}
