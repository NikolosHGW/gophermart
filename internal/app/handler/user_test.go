package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const (
	existLogin      = "user1"
	validToken      = "abc"
	correctLogin    = "user"
	correctPassword = "abc"
)

type MockUserService struct{}

func (m *MockUserService) Register(ctx context.Context, login, password string) (*entity.User, error) {
	if login == existLogin {
		return nil, domain.ErrLoginAlreadyExists
	}

	return &entity.User{
		ID:       1,
		Login:    login,
		Password: password,
	}, nil
}

func (m *MockUserService) GenerateJWT(user *entity.User) (string, error) {
	return validToken, nil
}

func (m *MockUserService) Authenticate(ctx context.Context, login, password string) (*entity.User, error) {
	if correctLogin == login && correctPassword == password {
		return &entity.User{
			ID:       1,
			Login:    login,
			Password: password,
		}, nil
	}

	return nil, domain.ErrInvalidCredentials
}

func TestUserHandler_RegisterUser(t *testing.T) {
	tests := []struct {
		name           string
		requestJSON    string
		returnUser     *entity.User
		returnJWT      string
		expectedStatus int
	}{
		{
			name:           "Положительный тест: успешная регистрация",
			requestJSON:    `{ "login": "new_user", "password": "abc" }`,
			returnJWT:      validToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Отрицательный тест: невалидные данные (логин число)",
			requestJSON:    `{ "login": 1, "password": "abc" }`,
			returnJWT:      validToken,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Отрицательный тест: невалидные данные (пароль число)",
			requestJSON:    `{ "login": "new_user", "password": 1 }`,
			returnJWT:      validToken,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Отрицательный тест: невалидные данные (логин отсутствует)",
			requestJSON:    `{ "status": "new_user", "password": "abc" }`,
			returnJWT:      validToken,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Отрицательный тест: невалидные данные (пароль отсутствует)",
			requestJSON:    `{ "login": "new_user", "status": "abc" }`,
			returnJWT:      validToken,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Отрицательный тест: логин уже занят",
			requestJSON:    `{ "login": "user1", "password": "abc" }`,
			returnJWT:      validToken,
			expectedStatus: http.StatusConflict,
		},
	}

	logger, _ := zap.NewDevelopment()
	s := &MockUserService{}
	h := NewUserHandler(s, logger)

	server := httptest.NewServer(http.HandlerFunc(h.RegisterUser))
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer([]byte(tt.requestJSON)))
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				authHeader := resp.Header.Get("Authorization")
				expectedAuthHeader := "Bearer " + validToken
				assert.Equal(t, expectedAuthHeader, authHeader)
			}

			err = resp.Body.Close()
			assert.NoError(t, err)
		})
	}
}

func TestUserHandler_LoginUser(t *testing.T) {
	tests := []struct {
		name           string
		requestJSON    string
		expectedStatus int
		expectedToken  string
	}{
		{
			name:           "Положительный тест: успешная аутентификация",
			requestJSON:    `{ "login": "user", "password": "abc" }`,
			expectedStatus: http.StatusOK,
			expectedToken:  validToken,
		},
		{
			name:           "Отрицательный тест: неверные учетные данные",
			requestJSON:    `{ "login": "user", "password": "wrong" }`,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Отрицательный тест: неверный формат запроса (нет ни логина, ни пароля)",
			requestJSON:    `{ "status": "user", "id": "abc" }`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	logger, _ := zap.NewDevelopment()
	s := &MockUserService{}
	h := NewUserHandler(s, logger)

	server := httptest.NewServer(http.HandlerFunc(h.LoginUser))
	defer server.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer([]byte(tt.requestJSON)))
			assert.NoError(t, err)

			resp, err := http.DefaultClient.Do(req)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.expectedStatus == http.StatusOK {
				authHeader := resp.Header.Get("Authorization")
				expectedAuthHeader := "Bearer " + tt.expectedToken
				assert.Equal(t, expectedAuthHeader, authHeader)
			}

			err = resp.Body.Close()
			assert.NoError(t, err)
		})
	}
}
