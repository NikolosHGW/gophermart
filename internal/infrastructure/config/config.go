package config

import (
	"flag"
	"fmt"
	"log"

	"github.com/caarlos0/env"
)

type config struct {
	RunAddress           string `env:"RUN_ADDRESS"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS"`
	SecretKey            string `env:"SECRET_KEY"`
}

func (c *config) InitEnv() error {
	err := env.Parse(c)
	if err != nil {
		return fmt.Errorf("не удалось разобрать уровень логирования: %w", err)
	}

	return nil
}

func (c *config) parseFlags() {
	flag.StringVar(&c.RunAddress, "a", "localhost:8080", "net address host:port")
	flag.StringVar(&c.DatabaseURI, "d",
		"user=nikolos "+
			"password=abc123 "+
			"dbname=gophermart "+
			"sslmode=disable",
		"data source name for connection")
	flag.StringVar(&c.AccrualSystemAddress, "r", "localhost:5000", "net address for Accrual System host:port")
	flag.StringVar(&c.SecretKey, "k", "abc", "secret key for hash")
	flag.Parse()
}

func NewConfig() *config {
	cfg := new(config)

	cfg.parseFlags()
	if err := cfg.InitEnv(); err != nil {
		log.Fatalf("Ошибка при инициализации переменных окружения: %v", err)
	}

	return cfg
}

func (c config) GetRunAddress() string {
	return c.RunAddress
}

func (c config) GetDatabaseURI() string {
	return c.DatabaseURI
}

func (c config) GetAccrualSystemAddress() string {
	return c.AccrualSystemAddress
}

func (c config) GetSecretKey() string {
	return c.SecretKey
}
