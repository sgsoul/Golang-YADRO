package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sgsoul/pkg/config"
	"github.com/sgsoul/pkg/database"
	sr "github.com/sgsoul/pkg/search"
	"github.com/sgsoul/pkg/words"
	"github.com/sgsoul/pkg/xkcd"
)

func main() {
	var (
		configPath     string
		searchString   string
		useIndex       bool
		relevantComics []database.Comic
	)

	flag.StringVar(&configPath, "c", "config.yaml", "path to configuration file")
	flag.StringVar(&searchString, "s", "", "serach string")
	flag.BoolVar(&useIndex, "i", false, "use index file for searching")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Println("error loading configuration fail:", err)
		return
	}

	if searchString != "" {
		db, err := database.LoadComicsFromFile(cfg.DBFile)
		if err != nil {
			fmt.Println("probably u should just run ./xkcd first\nerror loading comics from database file:", err)
		}

		normalizedKeywords := words.NormalizeWords(searchString)

		if useIndex {
			sr.BuildIndex(db, "index.json")
			index, err := os.ReadFile(cfg.IndexFile)
			if err != nil {
				fmt.Println("error loading index fail:", err)
				return
			}

			// поиск комиксов по индексу
			relevantComics = sr.RelevantComic(sr.IndexSearch(index, normalizedKeywords), cfg.DBFile)

		} else {
			// поиск комиксов по бд
			relevantComics = sr.FindRelevantComics(db, normalizedKeywords)
		}

		fmt.Println("Most relevant comics:")
		for i, comic := range relevantComics {
			if i >= 10 {
				break
			}
			fmt.Printf("Comic %d: %s\n", i+1, comic.URL)
		}

	} else {
		client := xkcd.NewClient(cfg.SourceURL)
		client.RunWorkers(cfg.Parallel, cfg.DBFile)
	}
}
