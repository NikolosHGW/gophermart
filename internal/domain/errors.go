package domain

import "errors"

var (
	ErrLoginAlreadyExists                = errors.New("логин уже существует")
	ErrInvalidCredentials                = errors.New("неверная пара логин/пароль")
	ErrOrderAlreadyUploadedByAnotherUser = errors.New("номер заказа уже был загружен другим пользователем")
	ErrOrderAlreadyUploadedForThisUser   = errors.New("номер заказа уже был загружен этим пользователем")
	ErrInternalServer                    = errors.New("внутренняя ошибка сервера")
	ErrAuth                              = errors.New("пользователь не авторизован")
)
