package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"go.uber.org/zap"
)

type OrderHandler struct {
	orderUseCase usecase.OrderUseCase
	logger       *zap.Logger
}

func NewOrderHandler(orderUseCase usecase.OrderUseCase, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		orderUseCase: orderUseCase,
		logger:       logger,
	}
}

func (h *OrderHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Info("не удалось прочитать тело запроса", zap.Error(err))
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	orderNumber := string(body)

	if !ValidateOrderNumber(orderNumber) {
		http.Error(w, "неверный формат номера заказа", http.StatusUnprocessableEntity)
		return
	}

	userID, ok := r.Context().Value(domain.ContextKey).(int)
	if !ok {
		h.logger.Info("userID не найден или неверного типа")
		http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	err = h.orderUseCase.ProcessOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrOrderAlreadyUploadedForThisUser):
			w.WriteHeader(http.StatusOK)
			return
		case errors.Is(err, domain.ErrOrderAlreadyUploadedByAnotherUser):
			http.Error(w, err.Error(), http.StatusConflict)
			return
		default:
			h.logger.Info("внутренняя ошибка сервера", zap.Error(err))
			http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusAccepted)
}

const lastDigit = 9

func ValidateOrderNumber(number string) bool {
	var sum int
	digits := make([]int, len(number))
	for i, char := range number {
		if char < '0' || char > '9' {
			return false
		}
		digits[i] = int(char - '0')
	}

	double := false
	for i := len(digits) - 1; i >= 0; i-- {
		digit := digits[i]
		if double {
			digit *= 2
			if digit > lastDigit {
				digit -= lastDigit
			}
		}
		sum += digit
		double = !double
	}

	return sum%10 == 0
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(domain.ContextKey).(int)
	if !ok {
		h.logger.Info("userID не найден или неверного типа")
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	orders, err := h.orderUseCase.GetUserOrdersByID(r.Context(), userID)
	if err != nil {
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	jsonResponse, err := json.Marshal(orders)
	if err != nil {
		h.logger.Info("ошибка при формировании ответа", zap.Error(err))
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		h.logger.Info("ошибка при отправки json", zap.Error(err))
		http.Error(w, domain.ErrInternalServer.Error(), http.StatusInternalServerError)
	}
}
