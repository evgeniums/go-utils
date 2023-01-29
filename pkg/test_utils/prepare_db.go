package test_utils

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/db/db_gorm"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var SqliteFolder string = ""

func SqlitePath(config *db.DBConfig) string {
	path := config.DB_EXTRA_CONFIG
	if SqliteFolder == "" {
		path = filepath.Join(os.TempDir(), path)
	} else {
		path = filepath.Join(SqliteFolder, path)
	}
	return path
}

func DbGormOpener(provider string, dsn string) (gorm.Dialector, error) {

	switch provider {
	case "postgres":
		return postgres.Open(dsn), nil
	case "sqlite":
		return sqlite.Open(dsn), nil
	}

	return nil, errors.New("unknown database provider")
}

func DbDsnBuilder(config *db.DBConfig) (string, error) {

	switch config.DB_PROVIDER {
	case "postgres":
		return db_gorm.PostgresDsnBuilder(config)
	case "sqlite":
		return SqlitePath(config), nil
	}

	return "", errors.New("unknown database provider")
}

func SetupGormDB() {
	db_gorm.DefaultDbConnector = func() *db_gorm.DbConnector {
		c := &db_gorm.DbConnector{}
		c.DialectorOpener = DbGormOpener
		c.DsnBuilder = DbDsnBuilder
		return c
	}
}
