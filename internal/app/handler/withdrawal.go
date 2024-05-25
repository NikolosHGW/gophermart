package handler

import (
	"encoding/json"
	"net/http"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"go.uber.org/zap"
)

type WithdrawalHandler struct {
	balanceUseCase    usecase.BalanceUseCase
	withdrawalUseCase usecase.WithdrawalUseCase
	orderUseCase      usecase.OrderUseCase
	logger            *zap.Logger
}

func NewWithdrawalHandler(
	balanceUseCase usecase.BalanceUseCase,
	withdrawalUseCase usecase.WithdrawalUseCase,
	orderUseCase usecase.OrderUseCase,
	logger *zap.Logger,
) *WithdrawalHandler {
	return &WithdrawalHandler{
		balanceUseCase:    balanceUseCase,
		withdrawalUseCase: withdrawalUseCase,
		orderUseCase:      orderUseCase,
		logger:            logger,
	}
}

type WithdrawRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *WithdrawalHandler) Withdraw(w http.ResponseWriter, r *http.Request) {
	var req WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Info("ошибки при декодинге request body", zap.Error(err))
		http.Error(w, "неверный номер заказа", http.StatusUnprocessableEntity)
		return
	}

	userID, ok := r.Context().Value(domain.ContextKey).(int)
	if !ok {
		http.Error(w, domain.ErrAuth.Error(), http.StatusUnauthorized)
		return
	}

	current, _, err := h.balanceUseCase.GetBalanceByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if !h.withdrawalUseCase.ValidBalance(current, req.Sum) {
		http.Error(w, "на счету недостаточно средств", http.StatusPaymentRequired)
		return
	}

	if !ValidateOrderNumber(req.Order) {
		http.Error(w, "неверный номер заказа", http.StatusUnprocessableEntity)
		return
	}

	if err := h.withdrawalUseCase.WithdrawFunds(r.Context(), userID, req.Order, req.Sum); err != nil {
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set(ContentType, ApplicationJSON)
	w.WriteHeader(http.StatusOK)

	err = r.Body.Close()
	if err != nil {
		h.logger.Info("ошибка при закрытии body", zap.Error(err))
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
	}
}

func (h *WithdrawalHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKey).(int)
	if !ok {
		http.Error(w, domain.ErrAuth.Error(), http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.withdrawalUseCase.GetWithdrawalsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set(ContentType, ApplicationJSON)

	resp, err := json.Marshal(withdrawals)
	if err != nil {
		h.logger.Info("ошибка при encoding списаний", zap.Error(err))
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}
	h.logger.Info("тело ответа", zap.Any("body", resp))
	_, err = w.Write(resp)
	if err != nil {
		h.logger.Info("ошибка при записи в тело ответа", zap.Error(err))
	}
}
