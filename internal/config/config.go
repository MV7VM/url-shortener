package config

import (
	"flag"
	"os"

	"github.com/gofrs/uuid"
)

func NewConfig() (*Model, error) {
	var cfg Model

	flag.StringVar(&cfg.HTTP.Host, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.HTTP.ReturningURL, "b", "http://localhost:8080/", "prefix of returning shart url")

	flag.StringVar(&cfg.Repo.SavingFilePath, "f", "./data.json", "file for recovery storage")
	flag.StringVar(&cfg.Repo.PsqlConnString, "d", "", "file for recovery storage")

	flag.StringVar(&cfg.Audit.AuditFilePath, "audit-file", "", "file for recovery storage")
	flag.StringVar(&cfg.Audit.AuditURL, "audit-url", "", "file for recovery storage")

	flag.Parse()

	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		cfg.Repo.SavingFilePath = filePath
	}

	if dbConn := os.Getenv("DATABASE_DSN"); dbConn != "" {
		cfg.Repo.SavingFilePath = dbConn
	}

	if filePath := os.Getenv("AUDIT_FILE"); filePath != "" {
		cfg.Audit.AuditFilePath = filePath
	}

	if url := os.Getenv("AUDIT_URL"); url != "" {
		cfg.Audit.AuditURL = url
	}

	secretKey, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	cfg.HTTP.SecretToken = secretKey.String()

	return &cfg, nil
}
