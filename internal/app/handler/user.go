package handler

import (
	"net/http"

	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
}

func NewUserHandler(userUseCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
	}
}

func (h UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	_, err := h.userUseCase.Register("hello", "parol")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte{})
	if err != nil {
		return
	}
}
