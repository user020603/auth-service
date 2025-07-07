package repositories

import (
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestTokenRepository_SetRefreshToken(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewTokenRepository(rdb)

	userID := uint(1)
	token := "refresh-token"
	expiresIn := time.Hour
	key := "refresh_token:user1"

	mock.ExpectSet(key, token, expiresIn).SetVal("OK")

	err := repo.SetRefreshToken(userID, token, expiresIn)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_GetRefreshToken(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewTokenRepository(rdb)

	userID := uint(2)
	token := "refresh-token-2"
	key := "refresh_token:user2"

	mock.ExpectGet(key).SetVal(token)

	got, err := repo.GetRefreshToken(userID)
	assert.NoError(t, err)
	assert.Equal(t, token, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_GetRefreshToken_NotFound(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewTokenRepository(rdb)

	userID := uint(3)
	key := "refresh_token:user3"

	mock.ExpectGet(key).RedisNil()

	got, err := repo.GetRefreshToken(userID)
	assert.Error(t, err)
	assert.Empty(t, got)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_DeleteRefreshToken(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewTokenRepository(rdb)

	userID := uint(4)
	key := "refresh_token:user4"

	mock.ExpectDel(key).SetVal(1)

	err := repo.DeleteRefreshToken(userID)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_SetRefreshToken_Error(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewTokenRepository(rdb)

	userID := uint(5)
	token := "token"
	expiresIn := time.Minute
	key := "refresh_token:user5"

	mock.ExpectSet(key, token, expiresIn).SetErr(errors.New("set error"))

	err := repo.SetRefreshToken(userID, token, expiresIn)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTokenRepository_DeleteRefreshToken_Error(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	repo := NewTokenRepository(rdb)

	userID := uint(6)
	key := "refresh_token:user6"

	mock.ExpectDel(key).SetErr(errors.New("del error"))

	err := repo.DeleteRefreshToken(userID)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
