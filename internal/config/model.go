package config

type Model struct {
	HTTP HTTPConfig `yaml:"HTTP"`
}

type HTTPConfig struct {
	Host         string
	ReturningUrl string
}
