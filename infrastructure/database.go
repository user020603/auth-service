package infrastructure

import (
	"fmt"
	"thanhnt208/vcs-sms/auth-service/configs"
	"thanhnt208/vcs-sms/auth-service/pkg/logger"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	db     *gorm.DB
	logger logger.ILogger
}

func NewDatabase(cfg *configs.Config, logger logger.ILogger) (IDatabase, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Failed to get database instance", "error", err)
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logger.Info("Connected to database", "host", cfg.DBHost, "port", cfg.DBPort)

	return &Database{
		db:     db,
		logger: logger,
	}, nil
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
