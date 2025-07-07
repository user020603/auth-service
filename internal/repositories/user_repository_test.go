package repositories

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"thanhnt208/vcs-sms/auth-service/internal/models"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{})
	require.NoError(t, err)

	cleanup := func() {
		sqlDB.Close()
	}

	return gormDB, mock, cleanup
}

func TestNewUserRepository(t *testing.T) {
	db, _, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	assert.NotNil(t, repo)
}

func TestUserRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	user := &models.User{
		ID:       1,
		Username: "johndoe",
		Password: "secret",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "users"`).
		WithArgs(
			user.Username,    // username
			user.Password,    // password
			sqlmock.AnyArg(), // name (empty string)
			sqlmock.AnyArg(), // email (empty string)
			sqlmock.AnyArg(), // role (empty string)
			user.ID,          // id
		).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(user.ID))
	mock.ExpectCommit()

	repo := NewUserRepository(db)
	err := repo.Create(user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByUsername(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "username", "password"}).
		AddRow(1, "johndoe", "secret")

	// GORM adds LIMIT as a second argument
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE username = .+ ORDER BY "users"\."id" LIMIT .+`).
		WithArgs("johndoe", 1).
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.FindByUsername("johndoe")

	assert.NoError(t, err)
	assert.Equal(t, "johndoe", user.Username)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "username", "password"}).
		AddRow(1, "johndoe", "secret")

	// GORM adds LIMIT as a second argument
	mock.ExpectQuery(`SELECT \* FROM "users" WHERE "users"\."id" = .+ ORDER BY "users"\."id" LIMIT .+`).
		WithArgs(uint(1), 1).
		WillReturnRows(rows)

	repo := NewUserRepository(db)
	user, err := repo.FindByID(1)

	assert.NoError(t, err)
	assert.Equal(t, uint(1), user.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}