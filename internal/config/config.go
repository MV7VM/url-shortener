package config

import (
	"flag"
	"os"
)

func NewConfig() (*Model, error) {
	var cfg Model

	flag.StringVar(&cfg.HTTP.Host, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.HTTP.ReturningURL, "b", "http://localhost:8080/", "prefix of returning shart url")

	flag.StringVar(&cfg.Repo.SavingFilePath, "f", "./data.json", "file for recovery storage")
	flag.StringVar(&cfg.Repo.PsqlConnString, "d", "postgresql://localhost:6132/postgres", "file for recovery storage")

	flag.Parse()

	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		cfg.Repo.SavingFilePath = filePath
	}

	if dbConn := os.Getenv("DATABASE_DSN"); dbConn != "" {
		cfg.Repo.SavingFilePath = dbConn
	}

	return &cfg, nil
}
