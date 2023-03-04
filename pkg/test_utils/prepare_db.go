package test_utils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/evgeniums/go-backend-helpers/pkg/app_context"
	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/evgeniums/go-backend-helpers/pkg/db/db_gorm"
	"github.com/evgeniums/go-backend-helpers/pkg/utils"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var SqliteFolder string = ""

func SqliteDatabasesPath() string {
	if SqliteFolder == "" {
		return filepath.Join(os.TempDir(), "go_sqlite_databases")
	}
	return SqliteFolder
}

func SqliteDbPath(fileName string) string {

	f := fileName
	if !strings.HasSuffix(f, ".sqlite") {
		f = utils.ConcatStrings(f, ".sqlite")
	}

	return filepath.Join(SqliteDatabasesPath(), f)
}

func SqlitePath(config *db.DBConfig) string {
	return SqliteDbPath(config.DB_NAME)
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

func DbCreator(provider string, db *gorm.DB, dbName string) error {

	switch provider {
	case "postgres":
		return db_gorm.PostgresDbCreator(provider, db, dbName)
	case "sqlite":
		return nil
	}

	return errors.New("unknown database provider")
}

func CheckDuplicateKeyError(provider string, result *gorm.DB) (bool, error) {

	switch provider {
	case "postgres":
		return db_gorm.PostgresCheckDuplicateKeyError(provider, result)
	case "sqlite":
		if err, ok := result.Error.(sqlite3.Error); ok {
			if err.ExtendedCode == sqlite3.ErrConstraintUnique {
				return true, errors.New("record already exists")

			}
		}
	}

	return false, result.Error
}

func SetupGormDB(t *testing.T) {
	db_gorm.NewModelStore(true)
	db_gorm.DefaultDbConnector = func() *db_gorm.DbConnector {
		c := &db_gorm.DbConnector{}
		c.DialectorOpener = DbGormOpener
		c.DsnBuilder = func(config *db.DBConfig) (string, error) {
			return DbDsnBuilder(t, config)
		}
		c.DbCreator = DbCreator
		return c
	}
}

func CreateDbModel(t *testing.T, app app_context.Context, models ...interface{}) {
	modelsList := append([]interface{}{}, models...)
	CreateDbModels(t, app, modelsList)
}

func CreateDbModels(t *testing.T, app app_context.Context, models []interface{}) {
	if models != nil {
		require.NoErrorf(t, app.Db().AutoMigrate(app, models), "failed to create database")
		db_gorm.GlobalModelStore.RegisterModels(models)
	}
}
