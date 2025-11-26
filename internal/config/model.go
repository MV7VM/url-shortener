package config

type Model struct {
	HTTP HTTPConfig `yaml:"HTTP"`
	Repo RepoConfig `yaml:"Repo"`
}

type HTTPConfig struct {
	Host         string
	ReturningURL string
}

type RepoConfig struct {
	CacheConfig
	PsqlConfig
}

type CacheConfig struct {
	SavingFilePath string
}

type PsqlConfig struct {
	PsqlConnString string
}
