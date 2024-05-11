package domain

import "errors"

var ErrLoginAlreadyExists = errors.New("логин уже существует")
var ErrInvalidCredentials = errors.New("неверная пара логин/пароль")
