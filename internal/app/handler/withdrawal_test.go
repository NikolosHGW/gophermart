package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockBalanceUseCaseForWithdrawal struct {
	mock.Mock
}

func (m *MockBalanceUseCaseForWithdrawal) GetBalanceByUserID(
	ctx context.Context,
	userID int,
) (float64, float64, error) {
	ret := m.Called(ctx, userID)
	return ret.Get(0).(float64), ret.Get(1).(float64), ret.Error(2)
}

type MockWithdrawalUseCase struct {
	mock.Mock
}

func (m *MockWithdrawalUseCase) ValidBalance(current, sum float64) bool {
	args := m.Called(current, sum)
	return args.Bool(0)
}

func (m *MockWithdrawalUseCase) WithdrawFunds(
	ctx context.Context,
	userID int,
	orderNumber string,
	sum float64,
) error {
	args := m.Called(ctx, userID, orderNumber, sum)
	return args.Error(0)
}

func (m *MockWithdrawalUseCase) GetWithdrawalsByUserID(
	ctx context.Context,
	userID int,
) ([]entity.Withdrawal, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.Withdrawal), args.Error(1)
}

type MockOrderUseCaseForWithdrawal struct {
	mock.Mock
}

func (m *MockOrderUseCaseForWithdrawal) ProcessOrder(ctx context.Context, userID int, orderNumber string) error {
	args := m.Called(ctx, userID, orderNumber)
	return args.Error(0)
}

func (m *MockOrderUseCaseForWithdrawal) GetUserOrdersByID(ctx context.Context, userID int) ([]entity.Order, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]entity.Order), args.Error(1)
}

func (m *MockOrderUseCaseForWithdrawal) OrderExists(
	ctx context.Context,
	userID int,
	orderNumber string,
) (bool, error) {
	args := m.Called(ctx, userID, orderNumber)
	return args.Bool(0), args.Error(1)
}

func TestWithdrawalHandler_Withdraw(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		request        WithdrawRequest
		userID         int
		setupMocks     func() *WithdrawalHandler
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Успешное списание средств",
			request: WithdrawRequest{
				Order: "2377225624",
				Sum:   100.0,
			},
			userID: 1,
			setupMocks: func() *WithdrawalHandler {
				balanceUseCase := new(MockBalanceUseCaseForWithdrawal)
				withdrawalUseCase := new(MockWithdrawalUseCase)
				orderUseCase := new(MockOrderUseCaseForWithdrawal)
				balanceUseCase.On("GetBalanceByUserID", mock.Anything, 1).Return(200.0, 0.0, nil)
				withdrawalUseCase.On("ValidBalance", 200.0, 100.0).Return(true)
				orderUseCase.On("OrderExists", mock.Anything, 1, "2377225624").Return(true, nil)
				withdrawalUseCase.On("WithdrawFunds", mock.Anything, 1, "2377225624", 100.0).Return(nil)

				return NewWithdrawalHandler(balanceUseCase, withdrawalUseCase, orderUseCase, logger)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "На счету недостаточно средств",
			request: WithdrawRequest{
				Order: "123456",
				Sum:   500.0,
			},
			userID: 1,
			setupMocks: func() *WithdrawalHandler {
				balanceUseCase := new(MockBalanceUseCaseForWithdrawal)
				withdrawalUseCase := new(MockWithdrawalUseCase)
				orderUseCase := new(MockOrderUseCaseForWithdrawal)
				balanceUseCase.On("GetBalanceByUserID", mock.Anything, 1).Return(300.0, 0.0, nil)
				withdrawalUseCase.On("ValidBalance", 300.0, 500.0).Return(false)

				return NewWithdrawalHandler(balanceUseCase, withdrawalUseCase, orderUseCase, logger)
			},
			expectedStatus: http.StatusPaymentRequired,
			expectedBody:   "на счету недостаточно средств\n",
		},
		{
			name: "Неверный номер заказа",
			request: WithdrawRequest{
				Order: "999999",
				Sum:   100.0,
			},
			userID: 1,
			setupMocks: func() *WithdrawalHandler {
				balanceUseCase := new(MockBalanceUseCaseForWithdrawal)
				withdrawalUseCase := new(MockWithdrawalUseCase)
				orderUseCase := new(MockOrderUseCaseForWithdrawal)
				balanceUseCase.On("GetBalanceByUserID", mock.Anything, 1).Return(200.0, 0.0, nil)
				withdrawalUseCase.On("ValidBalance", 200.0, 100.0).Return(true)
				orderUseCase.On("OrderExists", mock.Anything, 1, "999999").Return(false, nil)

				return NewWithdrawalHandler(balanceUseCase, withdrawalUseCase, orderUseCase, logger)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   "неверный номер заказа\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMocks()
			handler := tt.setupMocks()

			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(http.MethodPost, "/withdraw", bytes.NewBuffer(body))
			req = req.WithContext(context.WithValue(req.Context(), domain.ContextKey, tt.userID))

			w := httptest.NewRecorder()

			handler.Withdraw(w, req)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, tt.expectedStatus, res.StatusCode)
			if tt.expectedBody != "" {
				responseBody, _ := io.ReadAll(res.Body)
				assert.Equal(t, tt.expectedBody, string(responseBody))
			}
		})
	}
}

func TestWithdrawalHandler_GetWithdrawals(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		mockReturn     []entity.Withdrawal
		mockError      error
		expectedStatus int
	}{
		{
			name:           "успешная обработка запроса",
			userID:         1,
			mockReturn:     []entity.Withdrawal{{Order: "2377225624", Sum: 500, ProcessedAt: time.Now().Format((time.RFC3339))}},
			mockError:      nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "нет ни одного списания",
			userID:         1,
			mockReturn:     []entity.Withdrawal{},
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "внутренняя ошибка сервера",
			userID:         1,
			mockReturn:     nil,
			mockError:      domain.ErrInternalServer,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(MockWithdrawalUseCase)
			mockUseCase.On("GetWithdrawalsByUserID", mock.Anything, tt.userID).Return(tt.mockReturn, tt.mockError)

			req, _ := http.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			req = req.WithContext(context.WithValue(req.Context(), domain.ContextKey, tt.userID))
			rr := httptest.NewRecorder()

			handler := NewWithdrawalHandler(
				&MockBalanceUseCase{},
				mockUseCase,
				&MockOrderUseCaseForWithdrawal{},
				zap.NewNop(),
			)
			handler.GetWithdrawals(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
