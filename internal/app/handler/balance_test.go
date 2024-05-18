package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockBalanceUseCase struct {
	mock.Mock
}

func (_m *MockBalanceUseCase) GetBalanceByUserID(ctx context.Context, userID int) (float64, float64, error) {
	ret := _m.Called(ctx, userID)
	return ret.Get(0).(float64), ret.Get(1).(float64), ret.Error(2)
}

func TestBalanceHandler_GetBalance(t *testing.T) {
	mockService := new(MockBalanceUseCase)
	mockService.On("GetBalanceByUserID", mock.AnythingOfType("*context.valueCtx"), 1).Return(500.5, 42.0, nil)

	logger, _ := zap.NewDevelopment()
	handler := NewBalanceHandler(mockService, logger)

	req, _ := http.NewRequest(http.MethodGet, "/api/user/balance", nil)
	ctx := context.WithValue(req.Context(), domain.ContextKey, 1)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	handler.GetBalance(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	expectedBody := `{"current":500.5,"withdrawn":42}`
	assert.JSONEq(t, expectedBody, rr.Body.String())
}
