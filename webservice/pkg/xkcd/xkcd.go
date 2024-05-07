package xkcd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	db "github.com/sgsoul/pkg/database"
	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/pkg/words"
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

	// грузим инфо о комиксе
	resp, err := http.Get(fmt.Sprintf("%s/%d/info.0.json", c.baseURL, num))
	if err != nil {
		return comic, err
	}
	defer resp.Body.Close()

	// проверка на 404
	if resp.StatusCode == http.StatusNotFound {
		return comic, fmt.Errorf("comic %d not found", num)
	}

	// читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return comic, err
	}

	// DEBUG
	//fmt.Println("Response body:", string(body))

	// декодируем JSON
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

	// нормализация
	keywords := words.NormalizeWords(comicInfo.Title, comicInfo.Transcript, comicInfo.Alt)

	comic.URL = comicInfo.Img
	comic.Keywords = keywords

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

func (c *Client) RunWorkers(workers int, dbFile string) {
	var (
		wg                             sync.WaitGroup
		idChannel                      = make(chan int)         // канал для передачи id комиксов
		comicsChannel                  = make(chan ComicWithID) // канал для передачи загруженных комиксов
		doneChannel                    = make(chan struct{})    // канал для сигнализации о завершении
		expectedComics, receivedComics int
	)

	// считваем последний комикс на сайте
	latestComic, _ := c.retrieveLatestComicNum()

	log.Info().Msg("Loading comics..")

	// receiver
	go func() {
		defer close(idChannel)
		for i := 1; i <= latestComic; i++ {
			if _, err := os.Stat(dbFile); os.IsNotExist(err) {
				if _, err := os.Create(dbFile); err != nil {
					log.Error().Err(err).Msg("error creating the database")
					return
				}
			}
			comicsMap, _, err := db.LoadComicsFromFile(dbFile)
			if err != nil {
				log.Error().Err(err).Msg("error loading comics from the database")
				return
			}
			//проверка что комикс есть в дб
			if _, ok := comicsMap[fmt.Sprintf("%d", i)]; !ok && i != 404 {
				idChannel <- i // в канал загрузки
				expectedComics++
			}
		}
	}()

	wg.Add(workers)

	// workers
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for id := range idChannel {
				comic, err := c.retrieveComic(id)
				if err != nil {
					continue
				}
				comicsChannel <- ComicWithID{ID: id, Comic: comic}
			}
		}()
	}

	// collector
	go func() {
		defer close(comicsChannel)
		for comicWithID := range comicsChannel {
			db.SaveComicsToDatabase(dbFile, comicWithID.Comic, comicWithID.ID)
			receivedComics++
			if receivedComics == expectedComics {
				doneChannel <- struct{}{} // отправляем сигнал о завершении работы
			}
		}
	}()

	// wait 4 all worker goroutines to finish
	go func() {
		wg.Wait()
		close(doneChannel) // закрываем канал ожидания при завершении работы
	}()

	// ожидаем завершения работы
	<-doneChannel

	log.Info().Msg("Finish loading.")
}
