package config

import (
	"flag"
	"os"
)

func NewConfig() (*Model, error) {
	var (
		cfg Model
		ok  bool
	)

	cfg.HTTP.Host, ok = os.LookupEnv("SERVER_ADDRESS")
	if !ok {
		flag.StringVar(&cfg.HTTP.Host, "a", "localhost:8080", "address and port to run server")
	}

	cfg.HTTP.ReturningURL, ok = os.LookupEnv("BASE_URL")
	if !ok {
		flag.StringVar(&cfg.HTTP.ReturningURL, "b", "http://localhost:8080/", "address and port to run server")
	}

	flag.Parse()

	return &cfg, nil
}
