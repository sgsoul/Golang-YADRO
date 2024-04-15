package database

import (
	"encoding/json"
	"os"

	"github.com/sgsoul/golang-YADRO/pkg/xkcd"
)

func SaveJSONToFile(data interface{}, filename string) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, jsonData, 0644)
}

func LoadJSONFromFile(filename string, v interface{}) error {
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, v)
}

func LoadComicsFromFile(filename string) (map[string]xkcd.Comic, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var comicsMap = make(map[string]xkcd.Comic)

	decoder := json.NewDecoder(file)
	for decoder.More() {
		var comicData map[string]xkcd.Comic
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
