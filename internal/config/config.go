package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const defaultMaxFetchConcurrency = 5

type Config struct {
	DbAddress           string
	FetchInterval       time.Duration
	MaxFetchConcurrency int
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
	fetchInterval, err := time.ParseDuration(fetchIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("fetch interval is not valid. Valid time units are ns, us (or µs), ms, s, m, h: %w", err)
	}

	maxFetchConcurrency := defaultMaxFetchConcurrency
	maxFetchConcurrencyStr, ok := os.LookupEnv("MAX_FETCH_CONCURRENCY")
	if ok {
		maxFetchConcurrency, err = strconv.Atoi(maxFetchConcurrencyStr)
		if err != nil {
			return nil, fmt.Errorf("max fetch concurrency should be int: %w", err)
		}
	}

	return &Config{DbAddress: dbAddress, FetchInterval: fetchInterval, MaxFetchConcurrency: maxFetchConcurrency}, nil
}
