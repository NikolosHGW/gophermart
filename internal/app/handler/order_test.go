package handler

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

const (
	acceptedNumber = "12345678903"
	okNumber       = "4324802833166747"
	conflictNumber = "9278923470"
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
