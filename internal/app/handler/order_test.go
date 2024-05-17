package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const (
	acceptedNumber = "12345678903"
	okNumber       = "4324802833166747"
	conflictNumber = "9278923470"
)

var (
	firstUploadedAt  = time.Now().Format(time.RFC3339)
	secondUploadedAt = time.Now().Add(-time.Hour).Format(time.RFC3339)
)

type MockOrderUseCase struct{}

func (m *MockOrderUseCase) ProcessOrder(ctx context.Context, userID int, orderNumber string) error {
	switch orderNumber {
	case acceptedNumber:
		return nil
	case okNumber:
		return domain.ErrOrderAlreadyUploadedForThisUser
	case conflictNumber:
		return domain.ErrOrderAlreadyUploadedByAnotherUser
	}
	return errors.New("internal server error")
}

func (m *MockOrderUseCase) GetUserOrdersByID(ctx context.Context, userID int) ([]entity.Order, error) {
	if userID == 1 {
		return []entity.Order{
			{
				Number:     "12345678903",
				Status:     "NEW",
				UploadedAt: firstUploadedAt,
			},
			{
				Number:     "12345678904",
				Status:     "PROCESSING",
				UploadedAt: secondUploadedAt,
			},
		}, nil
	}
	return nil, domain.ErrInternalServer
}

func TestOrderHandler_UploadOrder(t *testing.T) {
	tests := []struct {
		name           string
		orderNumber    string
		userID         int
		expectedStatus int
	}{
		{
			name:           "Положительный тест: новый номер заказа",
			orderNumber:    acceptedNumber,
			userID:         1,
			expectedStatus: http.StatusAccepted,
		},
		{
			name:           "Отрицательный тест: номер заказа уже загружен этим пользователем",
			orderNumber:    okNumber,
			userID:         1,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Отрицательный тест: номер заказа уже загружен другим пользователем",
			orderNumber:    conflictNumber,
			userID:         2,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "Отрицательный тест: неверный формат номера заказа",
			orderNumber:    "invalid",
			userID:         1,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "Отрицательный тест: внутренняя ошибка сервера",
			orderNumber:    "",
			userID:         1,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	logger, _ := zap.NewDevelopment()
	mockUseCase := &MockOrderUseCase{}
	h := NewOrderHandler(mockUseCase, logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(
				http.MethodPost,
				"/api/user/orders",
				bytes.NewBuffer([]byte(tt.orderNumber)),
			)
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), domain.ContextKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			h.UploadOrder(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

func TestOrderHandler_GetOrders(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Положительный тест: список заказов",
			userID:         1,
			expectedStatus: http.StatusOK,
			expectedBody: fmt.Sprintf(
				`[{"number":"12345678903","status":"NEW","uploaded_at":"%s"},
				{"number":"12345678904","status":"PROCESSING","uploaded_at":"%s"}]`,
				firstUploadedAt,
				secondUploadedAt,
			),
		},
		{
			name:           "Отрицательный тест: внутренняя ошибка сервера",
			userID:         2,
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
	}

	logger, _ := zap.NewDevelopment()
	mockUseCase := &MockOrderUseCase{}
	h := NewOrderHandler(mockUseCase, logger)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/api/user/orders", nil)
			assert.NoError(t, err)

			ctx := context.WithValue(req.Context(), domain.ContextKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			h.GetOrders(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedBody != "" {
				var got []entity.Order
				err := json.Unmarshal(rr.Body.Bytes(), &got)
				assert.NoError(t, err)

				var want []entity.Order
				err = json.Unmarshal([]byte(tt.expectedBody), &want)
				assert.NoError(t, err)

				assert.Equal(t, want, got)
			}
		})
	}
}
