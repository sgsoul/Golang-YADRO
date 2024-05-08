package main

import (
	"flag"
	"fmt"

	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/config"
	"github.com/sgsoul/internal/core/database"
	sr "github.com/sgsoul/internal/util/search"
	"github.com/sgsoul/internal/util/words"
	"github.com/sgsoul/internal/adapters/xkcd"
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

	cfg := config.New(configPath)

	if searchString != "" {
		db := database.New(cfg.DBFile)

		normalizedKeywords := words.NormalizeWords(searchString)

		if useIndex {
			index := sr.New(db, cfg.IndexFile)
			// поиск комиксов по индексу
			relevantComics = sr.RelevantComic(sr.IndexSearch(index, normalizedKeywords), cfg.DBFile)

		} else {
			// поиск комиксов по бд
			relevantComics = sr.FindRelevantComics(db, normalizedKeywords)
		}

		log.Info().Msg("Most relevant comics:")
		for i, comic := range relevantComics {
			if i >= 10 {
				break
			}
			str := fmt.Sprintf("Comic %d: %s\n", i+1, comic.URL)
			log.Info().Msg(str)
		}

	} else {
		client := xkcd.NewClient(cfg.SourceURL)
		client.RunWorkers(cfg.Parallel, cfg.DBFile)
	}
}
