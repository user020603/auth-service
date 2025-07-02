package infrastructure

import (
	"context"
	"database/sql"
	"testing"
	"thanhnt208/vcs-sms/auth-service/config"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func mockGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, *sql.DB) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	// Set up ping expectation before gorm.Open, since GORM may ping on open
	mock.ExpectPing().WillReturnError(nil)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})

	require.NoError(t, err)

	// Set up ping expectation for the test's explicit db.Ping()
	mock.ExpectPing().WillReturnError(nil)
	// Set up close expectation for the test's explicit db.Close()
	mock.ExpectClose()
	return gormDB, mock, sqlDB
}

func TestDatabase_GetDB_Ping_Close(t *testing.T) {
	gdb, mock, raw := mockGormDB(t)
	defer raw.Close()

	db := NewTestDatabase(gdb)

	assert.Equal(t, gdb, db.GetDB())

	err := db.Ping(context.Background())
	assert.NoError(t, err)

	err = db.Close()
	assert.NoError(t, err)

	assert.NoError(t, mock.ExpectationsWereMet(), "there were unfulfilled expectations")
}

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