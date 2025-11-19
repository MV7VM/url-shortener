package config

import "flag"

func NewConfig() (*Model, error) {
	var cfg Model

	flag.StringVar(&cfg.HTTP.Host, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.HTTP.ReturningURL, "b", "", "address and port to run server")

	flag.Parse()

	return &cfg, nil
}
