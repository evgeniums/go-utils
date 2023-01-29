package test_utils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/db/db_gorm"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var SqliteFolder string = ""

func SqliteDbPath(fileName string) string {
	if SqliteFolder == "" {
		return filepath.Join(os.TempDir(), fileName)
	}
	return filepath.Join(SqliteFolder, fileName)
}

func SqlitePath(config *db.DBConfig) string {
	return SqliteDbPath(config.DB_EXTRA_CONFIG)
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

func DbDsnBuilder(t *testing.T, config *db.DBConfig) (string, error) {

	switch config.DB_PROVIDER {
	case "postgres":
		return db_gorm.PostgresDsnBuilder(config)
	case "sqlite":
		dsn := SqlitePath(config)
		t.Logf("Sqlite database DSN: %s", dsn)
		return dsn, nil
	}

	return "", errors.New("unknown database provider")
}

func SetupGormDB(t *testing.T) {
	db_gorm.DefaultDbConnector = func() *db_gorm.DbConnector {
		c := &db_gorm.DbConnector{}
		c.DialectorOpener = DbGormOpener
		c.DsnBuilder = func(config *db.DBConfig) (string, error) {
			return DbDsnBuilder(t, config)
		}
		return c
	}
}
