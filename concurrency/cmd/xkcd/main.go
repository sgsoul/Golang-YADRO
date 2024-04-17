package main

import (
	"flag"
	"fmt"

	"github.com/sgsoul/pkg/config"
	"github.com/sgsoul/pkg/database"
	"github.com/sgsoul/pkg/style"
	"github.com/sgsoul/pkg/xkcd"
)

func main() {
	var (
		outputFile bool
		maxComics  int
		configPath string
	)

	flag.BoolVar(&outputFile, "o", false, "print output to stdout")
	flag.IntVar(&maxComics, "n", -1, "max number of comics to retrieve")
	flag.StringVar(&configPath, "c", "config.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Println("error loading configuration fail:", err)
		return
	}
	filename := cfg.DBFile

	client := xkcd.NewClient(cfg.SourceURL)

	//проверка на -о флаг
	if outputFile {
		if maxComics < 0 {
			// выгрузка комиксов из дб
			comicsMap, err := database.LoadComicsFromFile(filename)
			if err != nil {
				fmt.Println("error loading comics from the database:", err)
				return
			}
			style.PrintAllComics(comicsMap)
		} else {
			// выгрузка комиксов из дб по макс значению из -n
			comicsMap, err := database.LoadComicsFromFile(filename)
			if err != nil {
				fmt.Println("error loading comics from the database:", err)
				return
			}
			style.PrintComics(comicsMap, maxComics)
		}
	} else {
		// если нет флага -о то скачиваем комиксы с сайта в дб
		client.RunWorkers(cfg.Parallel, filename)
	}
}
