package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	log "github.com/rs/zerolog/log"
	"github.com/sgsoul/internal/core"
	mocks "github.com/sgsoul/internal/server/mocks"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/time/rate"
)

func TestHandlePics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockSearch := mocks.NewMockSearch(ctrl)

	s := Server{
		config:  &core.Config{},
		service: mockService,
		search:  mockSearch,
	}

	mockSearch.EXPECT().RelevantURLS(gomock.Any(), gomock.Any()).Return([]string{"url1", "url2"}, nil)
	mockService.EXPECT().PrettyPrintService(gomock.Any()).Return(bytes.Buffer{})

	req, err := http.NewRequest("GET", "/pics?search=searchString", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.handlePics)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestHandleUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)

	s := Server{
		config:  &core.Config{},
		service: mockService,
	}

	req, err := http.NewRequest("POST", "/update", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.handleUpdate)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

func TestHandleLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)

	s := &Server{
		service: mockService,
		config: &core.Config{
			TokenTime: 15, // время жизни токена в минутах
		},
	}

	tests := []struct {
		name         string
		username     string
		password     string
		mockUser     core.User
		mockError    error
		expectedCode int
	}{
		{
			name:     "Валидные учетные данные",
			username: "testuser",
			password: "testpass",
			mockUser: core.User{
				Username: "testuser",
				Password: hashPassword("testpass"),
				Role:     "user",
			},
			mockError:    nil,
			expectedCode: http.StatusOK,
		},
		{
			name:         "Неверное имя пользователя",
			username:     "wronguser",
			password:     "testpass",
			mockError:    errors.New("user not found"),
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:     "Неверный пароль",
			username: "testuser",
			password: "wrongpass",
			mockUser: core.User{
				Username: "testuser",
				Password: hashPassword("testpass"),
				Role:     "user",
			},
			mockError:    nil,
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.EXPECT().
				Decode(gomock.Any(), gomock.Any(), gomock.Any()).
				DoAndReturn(func(w http.ResponseWriter, r *http.Request, v any) {
					*v.(*struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}) = struct {
						Username string `json:"username"`
						Password string `json:"password"`
					}{
						Username: tt.username,
						Password: tt.password,
					}
				})

			if tt.mockError != nil {
				mockService.EXPECT().
					GetUserByUsernameService(tt.username).
					Return(core.User{}, tt.mockError)
			} else {
				mockService.EXPECT().
					GetUserByUsernameService(tt.username).
					Return(tt.mockUser, nil)
			}

			reqBody, _ := json.Marshal(map[string]string{
				"username": tt.username,
				"password": tt.password,
			})

			req, err := http.NewRequest("POST", "/login", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()

			handler := http.HandlerFunc(s.handleLogin)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)

			if tt.expectedCode == http.StatusOK {
				assert.NotEmpty(t, rr.Header().Get("Set-Cookie"))
			} else {
				assert.Contains(t, rr.Body.String(), "Invalid username or password")
			}
		})
	}
}

func hashPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword)
}

func TestHandleRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)

	s := Server{
		config:  &core.Config{},
		service: mockService,
	}

	mockService.EXPECT().Decode(gomock.Any(), gomock.Any(), gomock.Any())
	mockService.EXPECT().CreateUserService(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	req, err := http.NewRequest("POST", "/register", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(s.handleRegister)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}
}

func TestStartServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockService := mocks.NewMockService(ctrl)
	mockSearch := mocks.NewMockSearch(ctrl)

	cfg := &core.Config{
		Port:      8080,
		TokenTime: 15,
		RateLim:   100,
		Parallel:  10,
		IndexFile: "index.json",
	}

	mockService.EXPECT().GetRateLimiter(gomock.Any(), gomock.Any()).Return(rate.NewLimiter(rate.Every(time.Second), 100)).AnyTimes()
	mockService.EXPECT().
    LimitedHandlerService(gomock.Any()).
    DoAndReturn(func(handler http.HandlerFunc) http.HandlerFunc {
        return handler
    }).
    AnyTimes()

	s, err := NewServer(cfg, mockService, mockSearch)
	assert.NoError(t, err)

	var logBuffer bytes.Buffer
	log.Logger = log.Output(&logBuffer)

	go func() {
		if err := s.Start(); err != nil {
			log.Error().Err(err).Msg("Error starting server")
		}
	}()


	time.Sleep(100 * time.Millisecond)

	assert.Empty(t, logBuffer.String(), "Unexpected error starting server")

	_, err = http.Get("http://localhost:8080")
	assert.NoError(t, err)
}
