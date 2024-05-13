package xkcd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	log "github.com/rs/zerolog/log"
	db "github.com/sgsoul/internal/core/database"
	"github.com/sgsoul/internal/util/words"
)

type Client struct {
	baseURL string
}

type ComicWithID struct {
	ID    int
	Comic db.Comic
}

func NewClient(url string) *Client {
	return &Client{
		baseURL: url,
	}
}

func (c *Client) retrieveComic(num int) (db.Comic, error) {
    var comic db.Comic

    // Загружаем информацию о комиксе
    resp, err := http.Get(fmt.Sprintf("%s/%d/info.0.json", c.baseURL, num))
    if err != nil {
        return comic, err
    }
    defer resp.Body.Close()

    // Проверяем наличие комикса по коду статуса HTTP
    if resp.StatusCode == http.StatusNotFound {
        return comic, fmt.Errorf("comic %d not found", num)
    }

    // Читаем тело ответа
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return comic, err
    }

    // Декодируем JSON
    var comicInfo struct {
        Img        string `json:"img"`
        Alt        string `json:"alt"`
        Transcript string `json:"transcript"`
        Title      string `json:"title"`
    }
    err = json.Unmarshal(body, &comicInfo)
    if err != nil {
        return comic, err
    }

    // Нормализуем ключевые слова
    keywords := words.NormalizeWords(comicInfo.Title, comicInfo.Transcript, comicInfo.Alt)

    // Преобразуем слайс ключевых слов в строку, разделенную запятыми
    keywordsStr := strings.Join(keywords, ",")

    // Заполняем структуру Comic
    comic.URL = comicInfo.Img
    comic.Keywords = keywordsStr

    return comic, nil
}



func (c *Client) retrieveLatestComicNum() (int, error) {
	resp, err := http.Get(fmt.Sprintf("%s/info.0.json", c.baseURL))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var info struct {
		Num int `json:"num"`
	}
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return 0, err
	}

	return info.Num, nil
}

func (c *Client) RunWorkers(workers int, dbvar *db.DB) {
    var (
        wg                             sync.WaitGroup
        comicsChannel                  = make(chan db.Comic)    // Channel for passing loaded comics
        doneChannel                    = make(chan struct{})    // Channel to signal completion
    )

    // Retrieve the latest comic number from the site
    latestComic, _ := c.retrieveLatestComicNum()

    log.Info().Msg("Loading comics..")

    wg.Add(1)

    // Worker for checking comic existence and loading comics
    go func() {
        defer wg.Done()

        for i := 1; i <= latestComic; i++ {
            // Check if the comic exists in the database
            _, err := dbvar.GetComicByID(i)
            if err != nil {
                // Comic does not exist, retrieve and save it
                comic, err := c.retrieveComic(i)
                if err != nil {
                    continue
                }
                comicsChannel <- comic // Send the comic to the channel for processing
            }
        }

        close(comicsChannel) // Close comicsChannel after all comics are sent
    }()

    wg.Add(workers)

    // Workers for processing comics
    for i := 0; i < workers; i++ {
        go func() {
            defer wg.Done()

            for comic := range comicsChannel {
                // Save the comic to the database
                if err := dbvar.SaveComicToDatabase(comic); err != nil {
                    continue
                }
            }
        }()
    }

    // Wait for completion
    go func() {
        wg.Wait() // Wait for all worker goroutines to finish
        doneChannel <- struct{}{} // Signal completion
    }()

    // Wait for completion
    <-doneChannel

    log.Info().Msg("Finished loading.")
}
