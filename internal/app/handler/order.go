package handler

import (
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
		h.logger.Error("не удалось прочитать тело запроса", zap.Error(err))
		http.Error(w, "неверный формат запроса", http.StatusBadRequest)
		return
	}
	orderNumber := string(body)

	if !ValidateOrderNumber(orderNumber) {
		http.Error(w, "неверный формат номера заказа", http.StatusUnprocessableEntity)
		return
	}

	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		h.logger.Error("userID не найден или неверного типа")
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
			h.logger.Error("внутренняя ошибка сервера", zap.Error(err))
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
