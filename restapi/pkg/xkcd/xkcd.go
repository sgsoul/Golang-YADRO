package xkcd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/sgsoul/golang-YADRO/pkg/words"
)

type Client struct {
	baseURL string
}

type Metadata struct {
	LastComic int `json:"last_comic"`
}

func NewClient() *Client {
	return &Client{
		baseURL: "https://xkcd.com",
	}
}

type Comic struct {
	URL      string   `json:"url"`
	Keywords []string `json:"keywords"`
}

func ReadMetadata(filename string) (*Metadata, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata Metadata
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// если файл пустой, создаем новую структуру метаданных и записываем ее в файл
	if fileInfo.Size() == 0 {
		metadata = Metadata{LastComic: 0}
		err = WriteMetadata(file, &metadata)
		if err != nil {
			return nil, err
		}
	} else {
		// иначе считываем содержимое файла метаданных
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&metadata)
		if err != nil {
			return nil, err
		}
	}

	return &metadata, nil
}

func WriteMetadata(file *os.File, metadata *Metadata) error {
	// передвигаем указатель в начало файла
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}

	// метаданные в JSON и пишем в файл
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(metadata); err != nil {
		return err
	}

	// DEBUG синхронизируем содержимое файла с диском (??)
	if err := file.Sync(); err != nil {
		return err
	}

	if err := file.Close(); err != nil {
		return err
	}

	return nil
}

func (c *Client) RetrieveComics(maxComics int, filename string) error {
	// считываем метадату
	metadata, err := ReadMetadata("metadata.json")
	if err != nil {
		return err
	}

	// считваем последний комикс на сайте
	latestComicNum, err := c.retrieveLatestComicNum()
	if err != nil {
		return err
	}

	// приводим ограничение по последнему комиксу
	if maxComics < 0 || maxComics > latestComicNum {
		maxComics = latestComicNum
	}

	// открываем дб файл
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// грузим комиксы в мапу
	// done == ограничение по обработке за раз. чтобы не банили =)
	done := 0
	for i := metadata.LastComic + 1; i <= metadata.LastComic+maxComics || maxComics < 0; i++ {
		if done >= 100 {
			break
		}

		comic, err := c.retrieveComic(i)
		if err != nil {
			// если нет комикса идем дальше
			if strings.Contains(err.Error(), "comic not found") {
				continue
			}
			// если следующий == выше последнего == поздраявлеям все скачали
			if strings.Contains(err.Error(), "comic "+fmt.Sprint(latestComicNum+1)+" not found") {
				fmt.Print("all comics have been successfully uploaded to the database")
			}
			// если не связано с тем что выше пишем об ошибке
			if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "comic "+fmt.Sprint(latestComicNum+1)) {
				return err
			}
			continue
		}

		// сериализуем и добавляем в дб
		comicJSON, err := json.MarshalIndent(map[string]Comic{fmt.Sprintf("%d", i): comic}, "", "   ")
		if err != nil {
			return err
		}
		_, err = file.Write(comicJSON)
		if err != nil {
			return err
		}
		_, err = file.WriteString("\n")
		if err != nil {
			return err
		}

		metadata.LastComic = i
		metadataFile, err := os.Create("metadata.json")
		if err != nil {
			return err
		}
		defer metadataFile.Close()

		err = WriteMetadata(metadataFile, metadata)
		if err != nil {
			return err
		}

		done++
	}

	return nil
}

func (c *Client) retrieveComic(num int) (Comic, error) {
	var comic Comic

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

func PrintComicsToStdout(comics []Comic) {
	for _, comic := range comics {
		fmt.Printf("URL: %s\n", comic.URL)
		fmt.Println("Keywords:", comic.Keywords)
		fmt.Println()
	}
}

func SaveComicsToFile(comics []Comic, filename string) error {
	data, err := json.MarshalIndent(comics, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}
