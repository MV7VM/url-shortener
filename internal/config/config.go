package config

import (
	"flag"
	"os"
)

func NewConfig() (*Model, error) {
	var (
		cfg Model
		//ok  bool
	)

	//godotenv.Load()

	cfg.HTTP.Host = os.Getenv("SERVER_ADDRESS")
	if cfg.HTTP.Host == "" {
		flag.StringVar(&cfg.HTTP.Host, "a", "localhost:8080", "address and port to run server")
	}

	cfg.HTTP.ReturningURL = os.Getenv("BASE_URL")
	if cfg.HTTP.ReturningURL == "" {
		flag.StringVar(&cfg.HTTP.ReturningURL, "b", "http://localhost:8080/", "address and port to run server")
	}

	return &cfg, nil
}
