package infrastructure

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"thanhnt208/vcs-sms/auth-service/config"
)

type Database struct {
	db *gorm.DB
}

var NewDatabase = func(cfg *config.Config) (IDatabase, error) {
	dsn := buildDSN(cfg)
	db, err := openGormDB(dsn, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return &Database{db: db}, nil
}

func NewTestDatabase(mockDB *gorm.DB) IDatabase {
	return &Database{db: mockDB}
}

func buildDSN(cfg *config.Config) string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBName,
		cfg.DBPort,
	)
}

func openGormDB(dsn string, gormCfg *gorm.Config) (*gorm.DB, error) {
	if gormCfg == nil {
		gormCfg = &gorm.Config{}
	}
	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func (d *Database) GetDB() *gorm.DB {
	return d.db
}

func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) Ping(ctx context.Context) error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}