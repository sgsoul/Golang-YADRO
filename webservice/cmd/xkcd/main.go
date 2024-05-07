package main

import (
	"flag"
	"fmt"
	"os"

	_ "github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
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
		log.Error().Err(err).Msg("error reading configuration fail")
		return
	}

	if searchString != "" {
		db, _,  err := database.LoadComicsFromFile(cfg.DBFile)
		if err != nil {
			log.Error().Err(err).Msg("probably u should just run ./xkcd first\nerror loading comics from database file")
		}

		normalizedKeywords := words.NormalizeWords(searchString)

		if useIndex {
			sr.BuildIndex(db, "index.json")
			index, err := os.ReadFile(cfg.IndexFile)
			if err != nil {
				log.Error().Err(err).Msg("error loading index fail")
				return
			}

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
		log.Info().Msg("all comics saved")
	}
}

