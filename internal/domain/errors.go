package domain

import "errors"

var ErrLoginAlreadyExists = errors.New("логин уже существует")
var ErrInvalidCredentials = errors.New("неверная пара логин/пароль")
var ErrOrderAlreadyUploadedByAnotherUser = errors.New("номер заказа уже был загружен другим пользователем")
var ErrOrderAlreadyUploadedForThisUser = errors.New("номер заказа уже был загружен этим пользователем")
