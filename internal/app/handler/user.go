package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/NikolosHGW/gophermart/internal/domain"
	"github.com/NikolosHGW/gophermart/internal/domain/entity"
	"github.com/NikolosHGW/gophermart/internal/domain/usecase"
	"go.uber.org/zap"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
	logger      *zap.Logger
}

func NewUserHandler(userUseCase usecase.UserUseCase, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		userUseCase: userUseCase,
		logger:      logger,
	}
}

func (h *UserHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	inputData, err := decodeAndValidateUserData(r, h.logger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userUseCase.Register(r.Context(), inputData.Login, inputData.Password)
	if err != nil {
		if errors.Is(err, domain.ErrLoginAlreadyExists) {
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendToken(w, h, user)
}

func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	inputData, err := decodeAndValidateUserData(r, h.logger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.userUseCase.Authenticate(r.Context(), inputData.Login, inputData.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sendToken(w, h, user)
}

type userData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func decodeAndValidateUserData(r *http.Request, logger *zap.Logger) (userData, error) {
	var data userData
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		logger.Info("ошибка декодирования: ", zap.Error(err))
		return data, fmt.Errorf("ошибка декодирования")
	}
	if data.Login == "" || data.Password == "" {
		return data, errors.New("неверный формат запроса")
	}
	return data, nil
}

func sendToken(w http.ResponseWriter, h *UserHandler, user *entity.User) {
	token, err := h.userUseCase.GenerateJWT(user)
	if err != nil {
		http.Error(w, "Ошибка при создании токена", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Authorization", "Bearer "+token)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
