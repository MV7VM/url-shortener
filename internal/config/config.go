// Package config содержит конфиг
// для сервиса сокращения ссылок.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/gofrs/uuid"
)

// fileConfig describes configuration fields that can be loaded from JSON file.
type fileConfig struct {
	ServerAddress string `json:"server_address"`
	BaseURL       string `json:"base_url"`
	FileStorage   string `json:"file_storage_path"`
	DatabaseDSN   string `json:"database_dsn"`
	EnableHTTPS   *bool  `json:"enable_https"`
	AuditFilePath string `json:"audit_file_path"`
	AuditURL      string `json:"audit_url"`
	TrustedSubnet string `json:"trusted_subnet"`
}

// NewConfig parses configuration in the following priority (low → high):
//  1. JSON file (if provided via -c/-config or CONFIG env)
//  2. Command-line flags
//  3. Environment variables
//
// A random HTTP secret token is generated on each application start.
func NewConfig() (*Model, error) {
	var (
		cfg      Model
		cfgPath  string
		jsonConf fileConfig
	)

	// Lowest priority: JSON file.
	flag.StringVar(&cfgPath, "c", "", "path to JSON config file")
	flag.StringVar(&cfgPath, "config", "", "path to JSON config file (alias)")

	// Defaults that are used when no other source provides a value.
	cfg.HTTP.Host = "localhost:8080"
	cfg.HTTP.ReturningURL = "http://localhost:8080/"
	cfg.Repo.SavingFilePath = "./data.json"
	cfg.Repo.PsqlConnString = ""
	cfg.HTTP.IsSecured = false

	// Read CONFIG env before flag.Parse, but allow flags to override it.
	if envPath := os.Getenv("CONFIG"); envPath != "" && cfgPath == "" {
		cfgPath = envPath
	}

	// Parse flags but do not bind directly into cfg, so that flags stay highest priority.
	hostFlag := flag.String("a", "", "address and port to run server")
	baseURLFlag := flag.String("b", "", "prefix of returning short url")
	fileStorageFlag := flag.String("f", "", "file for recovery storage")
	dsnFlag := flag.String("d", "", "database DSN")
	enableHTTPSFlag := flag.Bool("s", false, "enable HTTPS")
	TrustedSubnet := flag.String("t", "", "Trusted Subnet")

	flag.StringVar(&jsonConf.AuditFilePath, "audit-file", "", "audit log file path")
	flag.StringVar(&jsonConf.AuditURL, "audit-url", "", "audit HTTP endpoint")

	flag.Parse()

	// (Re-)apply CONFIG env over flags-defined cfgPath only if still empty.
	if cfgPath == "" {
		if envPath := os.Getenv("CONFIG"); envPath != "" {
			cfgPath = envPath
		}
	}

	// Load JSON config if path specified.
	if cfgPath != "" {
		if err := loadFileConfig(cfgPath, &jsonConf); err != nil {
			return nil, err
		}
	}

	// Apply JSON file (lowest priority).
	if jsonConf.ServerAddress != "" {
		cfg.HTTP.Host = jsonConf.ServerAddress
	}
	if jsonConf.BaseURL != "" {
		cfg.HTTP.ReturningURL = jsonConf.BaseURL
	}
	if jsonConf.FileStorage != "" {
		cfg.Repo.SavingFilePath = jsonConf.FileStorage
	}
	if jsonConf.DatabaseDSN != "" {
		cfg.Repo.PsqlConnString = jsonConf.DatabaseDSN
	}
	if jsonConf.EnableHTTPS != nil {
		cfg.HTTP.IsSecured = *jsonConf.EnableHTTPS
	}
	if jsonConf.AuditFilePath != "" {
		cfg.Audit.AuditFilePath = jsonConf.AuditFilePath
	}
	if jsonConf.AuditURL != "" {
		cfg.Audit.AuditURL = jsonConf.AuditURL
	}
	if jsonConf.TrustedSubnet != "" {
		cfg.HTTP.TrustedSubnet = jsonConf.TrustedSubnet
	}

	// Flags override JSON.
	if *hostFlag != "" {
		cfg.HTTP.Host = *hostFlag
	}
	if *baseURLFlag != "" {
		cfg.HTTP.ReturningURL = *baseURLFlag
	}
	if *fileStorageFlag != "" {
		cfg.Repo.SavingFilePath = *fileStorageFlag
	}
	if *dsnFlag != "" {
		cfg.Repo.PsqlConnString = *dsnFlag
	}
	if enableHTTPSFlag != nil && *enableHTTPSFlag {
		cfg.HTTP.IsSecured = true
	}
	if *TrustedSubnet != "" {
		cfg.HTTP.TrustedSubnet = *TrustedSubnet
	}

	// Environment variables have highest priority.
	if v := os.Getenv("SERVER_ADDRESS"); v != "" {
		cfg.HTTP.Host = v
	}
	if v := os.Getenv("BASE_URL"); v != "" {
		cfg.HTTP.ReturningURL = v
	}
	if v := os.Getenv("FILE_STORAGE_PATH"); v != "" {
		cfg.Repo.SavingFilePath = v
	}
	if v := os.Getenv("DATABASE_DSN"); v != "" {
		cfg.Repo.PsqlConnString = v
	}
	if v := os.Getenv("AUDIT_FILE"); v != "" {
		cfg.Audit.AuditFilePath = v
	}
	if v := os.Getenv("AUDIT_URL"); v != "" {
		cfg.Audit.AuditURL = v
	}
	if _, ok := os.LookupEnv("ENABLE_HTTPS"); ok {
		cfg.HTTP.IsSecured = true
	}
	if v := os.Getenv("TRUSTED_SUBNET"); v != "" {
		cfg.HTTP.TrustedSubnet = v
	}

	secretKey, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	cfg.HTTP.SecretToken = secretKey.String()

	return &cfg, nil
}

func loadFileConfig(path string, dst *fileConfig) error {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file: %w", err)
	}
	if len(data) == 0 {
		return errors.New("config file is empty")
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("unmarshal config file: %w", err)
	}
	return nil
}
