package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/NikolosHGW/gophermart/internal/domain"
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
	inputData, err := decodeAndValidateUserData(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = h.userUseCase.Register(r.Context(), inputData.Login, inputData.Password)
	if err != nil {
		if errors.Is(err, domain.ErrLoginAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

type userData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func decodeAndValidateUserData(r *http.Request) (userData, error) {
	var data userData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return data, fmt.Errorf("ошибка декодирования: %w", err)
	}
	if data.Login == "" || data.Password == "" {
		return data, errors.New("неверный формат запроса")
	}
	return data, nil
}
