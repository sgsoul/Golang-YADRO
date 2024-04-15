package main

import (
	"flag"
	"fmt"

	"github.com/sgsoul/golang-YADRO/pkg/database"
	"github.com/sgsoul/golang-YADRO/pkg/xkcd"
)

func main() {
	outputFile := flag.Bool("o", false, "print output to stdout")
	maxComics := flag.Int("n", -1, "max number of comics to retrieve")
	flag.Parse()

	client := xkcd.NewClient()

	// проверка на -о флаг
	if *outputFile {
		if *maxComics < 0 {
			// выгрузка комиксов из дб
			comicsMap, err := database.LoadComicsFromFile("database.json")
			if err != nil {
				fmt.Println("error loading comics from the database:", err)
				return
			}
			printAllComics(comicsMap)
		} else {
			// выгрузка комиксов из дб по макс значению из -n
			comicsMap, err := database.LoadComicsFromFile("database.json")
			if err != nil {
				fmt.Println("error loading comics from the database:", err)
				return
			}
			printComics(comicsMap, *maxComics)
		}
	} else {
		// если нет флага -о то скачиваем комиксы с сайта в дб
		filename := "database.json"
		if *maxComics < 0 {
			// все комиксы вообще
			if err := client.RetrieveComics(-1, filename); err != nil {
				fmt.Println("error retrieving comics:", err)
				return
			}
		} else {
			// или по ограничению
			if err := client.RetrieveComics(*maxComics, filename); err != nil {
				fmt.Println("error retrieving comics:", err)
				return
			}
		}
		fmt.Println("comics saved to", filename)
	}
}

// для принта с ограничением в колве
func printComics(comicsMap map[string]xkcd.Comic, maxComics int) {
	//fmt.Println("DEBUG printing comics with maxComics:", maxComics)
	i := 0
	for key, comic := range comicsMap {
		if i >= maxComics {
			break
		}
		printComic(key, comic)
		i++
	}
}

// для принта всех комиксов
func printAllComics(comicsMap map[string]xkcd.Comic) {
	for key, comic := range comicsMap {
		printComic(key, comic)
	}
}

// для красивого вывода
func printComic(key string, comic xkcd.Comic) {
	fmt.Printf("ID: %s\n", key)
	fmt.Printf("  image url: %s\n", comic.URL)
	fmt.Printf("  key words: ")
	for i, keyword := range comic.Keywords {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s", keyword)
	}
	fmt.Print("\n")
	fmt.Println()
}
