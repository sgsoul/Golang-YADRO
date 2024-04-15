package xkcd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/sgsoul/pkg/database"
	"github.com/sgsoul/pkg/style"
	"github.com/sgsoul/pkg/words"
)

type Client struct {
	baseURL string
}

type ComicWithID struct {
	ID    int
	Comic database.Comic
}

func NewClient(url string) *Client {
	return &Client{
		baseURL: url,
	}
}

func (c *Client) retrieveComic(num int) (database.Comic, error) {
	var comic database.Comic

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

func (c *Client) RunWorkers(workers int, dbFile string) {
	var (
		wg            sync.WaitGroup
		idChannel     = make(chan int)         // канал для передачи id комиксов
		comicsChannel = make(chan ComicWithID) // канал для передачи загруженных комиксов
	)

	loadingIndicator := style.StartLoadingIndicator()

	// receiver
	go func() {
		for i := 1; ; i++ {
			if _, err := os.Stat(dbFile); os.IsNotExist(err) {
				if _, err := os.Create(dbFile); err != nil {
					fmt.Println("error creating the database:", err)
					return
				}
			}
			comicsMap, err := database.LoadComicsFromFile(dbFile)
			if err != nil {
				fmt.Println("error loading comics from the database:", err)
				return
			}
			//проверка что комикс есть в дб
			if _, ok := comicsMap[fmt.Sprintf("%d", i)]; !ok && i != 404 {
				idChannel <- i // в канал загрузки
			}
		}
	}()

	// workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range idChannel {
				comic, err := c.retrieveComic(id)
				if err != nil {
					// check the next one
					_, err := c.retrieveComic(id + 1)
					if err != nil {
						// exit if double 404
						return
					}
					continue
				}
				comicsChannel <- ComicWithID{ID: id, Comic: comic}
			}
		}()
	}

	// collector
	go func() {
		defer close(comicsChannel) // закрываем канал при функции
		for comicWithID := range comicsChannel {
			database.SaveComicsToDatabase(dbFile, comicWithID.Comic, comicWithID.ID)
		}
	}()

	// wait 4 all worker goroutines to finish
	wg.Wait()
	style.StopLoadingIndicator(loadingIndicator)
	fmt.Println("\nall comics saved to", dbFile)
}
