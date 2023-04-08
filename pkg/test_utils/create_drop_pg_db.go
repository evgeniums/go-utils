package test_utils

import (
	"context"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

type PostgresDbConfig struct {
	DbHost     string
	DbPort     uint16
	DbUser     string
	DbPassword string

	DbNameForLogin string
}

func NewPostgresDbConfig() *PostgresDbConfig {
	p := &PostgresDbConfig{}
	p.DbHost = "127.0.0.1"
	p.DbPort = 5432
	p.DbUser = "bhelpers_user"
	p.DbPassword = "bhelpers_password"
	p.DbNameForLogin = "bhelpers_db_login"
	return p
}

func (p *PostgresDbConfig) Url() string {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", p.DbUser, p.DbPassword, p.DbHost, p.DbPort, p.DbNameForLogin)
	return dbUrl
}

func DropDatabase(t *testing.T, dbConfig *PostgresDbConfig, dbName string) {

	conn, err := pgx.Connect(context.Background(), dbConfig.Url())
	require.NoError(t, err)
	defer conn.Close(context.Background())

	_, err = conn.Query(context.Background(), fmt.Sprintf("DROP DATABASE %s WITH (FORCE);", dbName))
	require.NoError(t, err)
}

func CreateDatabase(t *testing.T, dbConfig *PostgresDbConfig, dbName string) {

	conn, err := pgx.Connect(context.Background(), dbConfig.Url())
	require.NoError(t, err)
	defer conn.Close(context.Background())

	_, err = conn.Query(context.Background(), fmt.Sprintf("CREATE DATABASE %s;", dbName))
	require.NoError(t, err)
}
