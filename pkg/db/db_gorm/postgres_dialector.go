package db_gorm

import (
	"errors"
	"fmt"

	"github.com/evgeniums/go-backend-helpers/pkg/db"
	"github.com/jackc/pgconn"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/jackc/pgerrcode"
)

func PostgresOpener(provider string, dsn string) (gorm.Dialector, error) {

	// switch provider {
	// case "postgres":
	return postgres.Open(dsn), nil
	// case "mysql":
	// 	return mysql.Open(dsn), nil
	// case "sqlite":
	// 	return sqlite.Open(dsn), nil
	// case "sqlserver":
	// 	return sqlserver.Open(dsn), nil
	// }

	// return nil, errors.New("unknown database provider")
}

func PostgresDsnBuilder(config *db.DBConfig) (string, error) {
	dsn := fmt.Sprintf("host=%v port=%v user=%v dbname=%v password=%v sslmode=disable", config.DB_HOST, config.DB_PORT, config.DB_USER, config.DB_NAME, config.DB_PASSWORD)
	return dsn, nil
}

func PostgresDbCreator(provider string, db *gorm.DB, dbName string) error {

	// check if db exists
	rs := db.Raw("SELECT * FROM pg_database WHERE datname = ?;", dbName)
	if rs.Error != nil {
		return fmt.Errorf("failed to select from pg_database: %s", rs.Error)
	}
	var rec = make(map[string]interface{})
	rs.Find(rec)

	// if not create it
	if len(rec) == 0 {
		// create database
		rs := db.Exec(fmt.Sprintf("CREATE DATABASE %s;", dbName))
		if rs.Error != nil {
			return fmt.Errorf("failed to create postgres database: %s", rs.Error)
		}
	}

	// done
	return nil
}

func PostgresCheckDuplicateKeyError(provider string, result *gorm.DB) (bool, error) {

	if err, ok := result.Error.(*pgconn.PgError); ok {
		if err.Code == pgerrcode.UniqueViolation {
			return true, errors.New("record already exists")
		}
	}

	return false, result.Error
}

func PostgresDbConnector() *DbConnector {
	c := &DbConnector{}
	c.DialectorOpener = PostgresOpener
	c.DsnBuilder = PostgresDsnBuilder
	c.CheckDuplicateKeyError = PostgresCheckDuplicateKeyError
	return c
}
