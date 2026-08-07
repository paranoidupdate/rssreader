package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DbAddress     string
	FetchInterval int
}

func InitConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}
	dbAddress := os.Getenv("DB_ADDRESS")
	if dbAddress == "" {
		return nil, errors.New("DB address is not defined")
	}
	fetchIntervalStr := os.Getenv("FETCH_INTERVAL")
	fetchInterval, err := strconv.Atoi(fetchIntervalStr) // TODO: use time.Duration
	if err != nil {
		return nil, fmt.Errorf("fetch interval is not int: %w", err)
	}

	return &Config{dbAddress, fetchInterval}, nil
}
