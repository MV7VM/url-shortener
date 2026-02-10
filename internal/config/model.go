package config

type Model struct {
	HTTP  HTTPConfig `yaml:"HTTP"`
	Repo  RepoConfig `yaml:"Repo"`
	Audit AuditorConfig
}

type HTTPConfig struct {
	Host         string
	ReturningURL string
	SecretToken  string
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

type AuditorConfig struct {
	AuditFilePath string
	AuditURL      string
}
