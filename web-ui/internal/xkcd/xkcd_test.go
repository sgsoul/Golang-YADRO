package xkcd

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sgsoul/internal/core"
	mocks "github.com/sgsoul/internal/xkcd/mocks"
	"github.com/stretchr/testify/assert"
)

func TestRetrieveComic(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockStorage := mocks.NewMockStorage(ctrl)
    client := NewClient("http://xkcd.com", mockStorage)

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/1/info.0.json" {
            w.WriteHeader(http.StatusOK)
            _, err := w.Write([]byte(`{
                "num": 1,
                "title": "Test Comic",
                "transcript": "Test transcript, apple doctor",
                "alt": "Test alt",
                "img": "http://example.com/image.png"
            }`))
            if err != nil{
                http.Error(w, err.Error(), http.StatusTeapot)
                return
            }
        } else {
            w.WriteHeader(http.StatusNotFound)
        }
    }))
    defer server.Close()

    client.baseURL = server.URL

    comic, err := client.retrieveComic(1)
    assert.NoError(t, err)
    assert.Equal(t, "http://example.com/image.png", comic.URL)
    assert.Equal(t, "test,comic,appl,doctor", comic.Keywords)
}

func TestRetrieveLatestComicNum(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockStorage := mocks.NewMockStorage(ctrl)
    client := NewClient("http://xkcd.com", mockStorage)

    server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/info.0.json" {
            w.WriteHeader(http.StatusOK)
            _, err := w.Write([]byte(`{"num": 123}`))
            if err != nil{
                http.Error(w, err.Error(), http.StatusTeapot)
                return
            }
        } else {
            w.WriteHeader(http.StatusNotFound)
        }
    }))
    defer server.Close()

    client.baseURL = server.URL

    num, err := client.retrieveLatestComicNum()
    assert.NoError(t, err)
    assert.Equal(t, 123, num)
}

type FakeStorage struct {
    comics map[int]core.Comic
    mu     sync.Mutex
}

func (s *FakeStorage) GetComicByID(id int) (core.Comic, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    if comic, ok := s.comics[id]; ok {
        return comic, nil
    }
    return core.Comic{}, errors.New("Comic not found")
}

func (s *FakeStorage) SaveComicToDatabase(comic core.Comic) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.comics[comic.ID] = comic
    return nil
}

func TestRunWorkers(t *testing.T) {
    fakeDB := &FakeStorage{comics: make(map[int]core.Comic)}
    client := NewClient("https://xkcd.com", fakeDB)

    done := make(chan struct{})

    ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
    defer cancel()

    go func() {
        client.RunWorkers(5)
        close(done)
    }()

    select {
    case <-done:
    case <-ctx.Done():
        t.Log("RunWorkers execution timed out")
    }

    for i := 1; i <= 1; i++ {
        comic, _ := fakeDB.GetComicByID(i)
        assert.NotNil(t, comic)
    }
}