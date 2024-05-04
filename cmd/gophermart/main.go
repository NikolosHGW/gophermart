package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/NikolosHGW/gophermart/internal/app/handler"
	"github.com/NikolosHGW/gophermart/internal/app/service"
	"github.com/NikolosHGW/gophermart/internal/infrastructure/config"
	"github.com/NikolosHGW/gophermart/internal/infrastructure/persistence"
	"github.com/NikolosHGW/gophermart/internal/infrastructure/persistence/db"
	"github.com/NikolosHGW/gophermart/internal/infrastructure/router"
	"github.com/NikolosHGW/gophermart/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(fmt.Errorf("не удалось запустить сервер: %w", err))
	}
}

func run() error {
	config := config.NewConfig()

	myLogger, err := logger.NewLogger("info")
	if err != nil {
		return fmt.Errorf("не удалось инициализировать логгер: %w", err)
	}

	database, err := db.InitDB(config.GetDatabaseURI())
	if err != nil {
		return fmt.Errorf("не удалось инициализировать базу данных: %w", err)
	}

	defer func() {
		if closeErr := database.Close(); closeErr != nil {
			myLogger.Fatal("ошибка при закрытии базы данных: ", zap.Error(err))
		}
	}()

	userRepo := persistence.NewSQLUserRepository(database)

	userService := service.NewUserService(userRepo)

	handlers := &handler.Handlers{
		UserHandler: handler.NewUserHandler(userService),
	}

	r := router.NewRouter(handlers)

	err = http.ListenAndServe(config.GetRunAddress(), r)

	return fmt.Errorf("ошибка при запуске сервера: %w", err)
}
