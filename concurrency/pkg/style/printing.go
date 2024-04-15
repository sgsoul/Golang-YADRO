package style

import (
	"fmt"

	"github.com/sgsoul/pkg/database"
)

// для принта с ограничением в колве
func PrintComics(comicsMap map[string]database.Comic, maxComics int) {
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
func PrintAllComics(comicsMap map[string]database.Comic) {
	for key, comic := range comicsMap {
		printComic(key, comic)
	}
}

// для красивого вывода
func printComic(key string, comic database.Comic) {
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