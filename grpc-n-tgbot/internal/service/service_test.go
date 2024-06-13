package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sgsoul/internal/core"
	mocks "github.com/sgsoul/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestUpdateDatabase(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockClient := mocks.NewMockClientXKCD(ctrl)

	comicsBefore := 5
	comicsAfter := 10

	mockStorage.EXPECT().GetCount().Return(comicsBefore, nil).Times(1)
	mockClient.EXPECT().RunWorkers(2).Times(1)
	mockStorage.EXPECT().GetCount().Return(comicsAfter, nil).Times(1)

	s := NewService(&core.Config{ConcLim: 10}, mockStorage, mockClient)

	result, err := s.UpdateDatabase(2)
	assert.NoError(t, err)
	assert.Equal(t, core.ComicCount{UpdatedComics: 5, TotalComics: 10}, result)
}

func TestPrettyPrintService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockClient := mocks.NewMockClientXKCD(ctrl)

	comics := []core.Comic{
		{ID: 1, URL: "test.xkcd/2", Keywords: "Spider-Man"},
		{ID: 2, URL: "test.xkcd/3", Keywords: "Batman"},
	}

	buffer := bytes.Buffer{}
	buffer.WriteString("Pretty Printed Comics")

	mockStorage.EXPECT().PrettyPrint(comics).Return(buffer)

	s := NewService(&core.Config{ConcLim: 10}, mockStorage, mockClient)

	result := s.PrettyPrintService(comics)
	assert.Equal(t, buffer, result)
}

func TestGetUserByUsernameService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockClient := mocks.NewMockClientXKCD(ctrl)

	username := "testuser"
	user := core.User{Username: "testuser", Role: "admin"}

	mockStorage.EXPECT().GetUserByUsername(username).Return(user, nil)

	s := NewService(&core.Config{ConcLim: 10}, mockStorage, mockClient)

	result, err := s.GetUserByUsernameService(username)
	assert.NoError(t, err)
	assert.Equal(t, user, result)
}

func TestCreateUserService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockClient := mocks.NewMockClientXKCD(ctrl)

	username := "newuser"
	password := "password"
	role := "user"

	mockStorage.EXPECT().CreateUser(username, password, role).Return(nil)

	s := NewService(&core.Config{ConcLim: 10}, mockStorage, mockClient)

	err := s.CreateUserService(username, password, role)
	assert.NoError(t, err)
}

func TestLimitedHandlerService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockClient := mocks.NewMockClientXKCD(ctrl)

	s := NewService(&core.Config{ConcLim: 2}, mockStorage, mockClient)

	handler := s.LimitedHandlerService(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req, _ := http.NewRequest("GET", "/", nil)
	w := &mockResponseWriter{}

	handler(w, req)
	assert.Equal(t, http.StatusOK, w.status)

	w = &mockResponseWriter{}
	handler(w, req)
	assert.Equal(t, http.StatusOK, w.status)
}

type mockResponseWriter struct {
	status int
}

func (m *mockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (m *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.status = statusCode
}

func TestInitConcurrencyLimiter(t *testing.T) {
	limit := int64(5)
	initConcurrencyLimiter(limit)
	assert.NotNil(t, Sem)
}

func SetupService(t *testing.T) *service{
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := mocks.NewMockStorage(ctrl)
	mockClient := mocks.NewMockClientXKCD(ctrl)

	s := NewService(&core.Config{}, mockStorage, mockClient)
	return s
}
func TestGetRateLimiter(t *testing.T) {
	s := SetupService(t)

	ip := "127.0.0.1"
	rps := 2
	limiter := s.GetRateLimiter(ip, rps)
	assert.NotNil(t, limiter)
	assert.Equal(t, rate.Limit(rps), limiter.Limit())
	assert.Equal(t, rps, limiter.Burst())

	limiter2 := s.GetRateLimiter(ip, rps)
	assert.Equal(t, limiter, limiter2)
}

func TestDecode(t *testing.T) {
	s := SetupService(t)

	type Payload struct {
		Name string `json:"name"`
	}

	payload := Payload{Name: "test"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	var result Payload
	s.Decode(w, req, &result)

	assert.Equal(t, payload.Name, result.Name)
	assert.Equal(t, http.StatusOK, w.Result().StatusCode)
}

func TestDecodeInvalidBody(t *testing.T) {
	s := SetupService(t)
	req := httptest.NewRequest("POST", "/", bytes.NewReader([]byte("invalid body")))
	w := httptest.NewRecorder()

	var result any
	s.Decode(w, req, &result)

	assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
}

func TestNew(t *testing.T) {
	s := SetupService(t)

	assert.NotNil(t, s, "should've been created..")
}
