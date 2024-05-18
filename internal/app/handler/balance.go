package handler

import (
	"encoding/json"
	"net/http"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"go.uber.org/zap"
)

type BalanceHandler struct {
	balanceUseCase usecase.BalanceUseCase
	logger         *zap.Logger
}

func NewBalanceHandler(balanceUseCase usecase.BalanceUseCase, logger *zap.Logger) *BalanceHandler {
	return &BalanceHandler{
		balanceUseCase: balanceUseCase,
		logger:         logger,
	}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKey).(int)
	if !ok {
		h.logger.Info("пользователь не авторизован")
		http.Error(w, "пользователь не авторизован", http.StatusUnauthorized)
		return
	}

	current, withdrawn, err := h.balanceUseCase.GetBalanceByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	balance := struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}{
		Current:   current,
		Withdrawn: withdrawn,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(balance)
	if err != nil {
		h.logger.Info("ошибка json encode", zap.Error(err))
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
	}
}
