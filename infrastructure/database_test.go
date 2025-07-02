package infrastructure

import (
	"context"
	"os"
	"testing"
	"thanhnt208/vcs-sms/auth-service/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewDatabase_Success(t *testing.T) {
	// Use SQLite in-memory DB for testing instead of real PostgreSQL
	mockDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	db := NewTestDatabase(mockDB)
	assert.NotNil(t, db.GetDB())
}

func TestNewDatabase_Failure(t *testing.T) {
	cfg := &config.Config{
		DBHost: "invalid", // force connection error
	}
	db, err := NewDatabase(cfg)
	assert.Nil(t, db)
	assert.Error(t, err)
}

func TestNewTestDatabase(t *testing.T) {
	mockDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	db := NewTestDatabase(mockDB)
	assert.NotNil(t, db.GetDB())
}

func TestBuildDSN(t *testing.T) {
	cfg := &config.Config{
		DBHost:     "localhost",
		DBUser:     "postgres",
		DBPassword: "password",
		DBName:     "dbname",
		DBPort:     "5432",
	}
	dsn := buildDSN(cfg)
	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "user=postgres")
	assert.Contains(t, dsn, "password=password")
	assert.Contains(t, dsn, "dbname=dbname")
	assert.Contains(t, dsn, "port=5432")
	assert.Contains(t, dsn, "sslmode=disable")
}

func TestOpenGormDB_Success(t *testing.T) {
	dsn := "file::memory:?cache=shared"
	db, err := openGormDB(dsn, &gorm.Config{SkipDefaultTransaction: true})
	// openGormDB uses postgres driver, so this should fail
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestOpenGormDB_Failure(t *testing.T) {
	dsn := "invalid_dsn"
	db, err := openGormDB(dsn, &gorm.Config{})
	assert.Error(t, err)
	assert.Nil(t, db)
}

// Integration test: requires a running Postgres container
// Set env vars or edit config as needed for your test DB
func TestIntegration_NewDatabase_Ping_Close(t *testing.T) {
	// Example: use environment variables for DB config
	cfg := &config.Config{
		DBHost:     getenvDefault("POSTGRES_HOST", "localhost"),
		DBUser:     getenvDefault("POSTGRES_USER", "postgres"),
		DBPassword: getenvDefault("POSTGRES_PASSWORD", "password"),
		DBName:     getenvDefault("POSTGRES_DB", "dbname"),
		DBPort:     getenvDefault("POSTGRES_PORT", "5432"),
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Fatalf("failed to connect to Postgres: %v", err)
	}
	defer db.Close()

	err = db.Ping(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, db.GetDB())
}

func getenvDefault(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// docker run --name postgres-test -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=password -e POSTGRES_DB=dbname -p 5432:5432 -d postgres:15
