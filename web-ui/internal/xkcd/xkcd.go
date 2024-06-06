package xkcd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/core"
	"github.com/sgsoul/internal/words"
)

//go:generate mockgen -source=xkcd.go -destination=mocks/mock.go

type Storage interface {
	GetComicByID(id int) (core.Comic, error)
	SaveComicToDatabase(comic core.Comic) error
}

type Client struct {
	baseURL string
	storage Storage
}

func NewClient(url string, st Storage) *Client {
	return &Client{
		baseURL: url,
		storage: st,
	}
}

func (c *Client) retrieveComic(num int) (core.Comic, error) {
	var comic core.Comic

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
	err = json.Unmarshal(body, &core.ComicInfo)
	if err != nil {
		return comic, err
	}

	// Нормализуем ключевые слова
	keywords := words.NormalizeWords(core.ComicInfo.Title, core.ComicInfo.Transcript, core.ComicInfo.Alt)

	// Преобразуем слайс ключевых слов в строку, разделенную запятыми
	keywordsStr := strings.Join(keywords, ",")

	// Заполняем структуру Comic
	comic.URL = core.ComicInfo.Img
	comic.Keywords = keywordsStr

	return comic, nil
}

func (c *Client) retrieveLatestComicNum() (int, error) {
	return c.retrieveLatestComicNumFromAPI()
}

func (c *Client) retrieveLatestComicNumFromAPI() (int, error) {
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

func (c *Client) RunWorkers(workers int) {
	var (
		wg            sync.WaitGroup
		comicsChannel = make(chan core.Comic)
		doneChannel   = make(chan struct{})
	)

	latestComic, _ := c.retrieveLatestComicNum()

	log.Info().Msg("Loading comics..")

	wg.Add(1)

	go func() {
		defer wg.Done()

		for i := 1; i <= latestComic; i++ {
			_, err := c.storage.GetComicByID(i)
			if err != nil {
				comic, err := c.retrieveComic(i)
				if err != nil {
					continue
				}
				comicsChannel <- comic
			}
		}

		close(comicsChannel)
	}()

	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			for comic := range comicsChannel {
				if err := c.storage.SaveComicToDatabase(comic); err != nil {
					continue
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		doneChannel <- struct{}{}
	}()

	<-doneChannel

	log.Info().Msg("Finished loading.")
}
