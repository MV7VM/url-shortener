package config

// Model aggregates all configuration sections used by the application.
type Model struct {
	HTTP  HTTPConfig `yaml:"HTTP"`
	Repo  RepoConfig `yaml:"Repo"`
	Audit AuditorConfig
}

// HTTPConfig contains network and HTTP-related settings.
type HTTPConfig struct {
	Host         string
	ReturningURL string
	SecretToken  string
}

// RepoConfig groups configuration for cache and PostgreSQL repositories.
type RepoConfig struct {
	CacheConfig
	PsqlConfig
}

// CacheConfig describes file-based cache storage settings.
type CacheConfig struct {
	SavingFilePath string
}

// PsqlConfig contains PostgreSQL connection configuration.
type PsqlConfig struct {
	PsqlConnString string
}

// AuditorConfig configures optional audit sinks such as file and HTTP endpoint.
type AuditorConfig struct {
	AuditFilePath string
	AuditURL      string
}
