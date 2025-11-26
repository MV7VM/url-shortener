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
	SavingFilePath string
}
