package persistence

import (
	"context"
	"fmt"

	"github.com/rhythin/sever-management/internal"
	"github.com/rhythin/sever-management/internal/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// NewDB initializes and migrates the database
func NewDB(cfg *internal.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if err := MigrateDB(ctx, db); err != nil {
		return nil, err
	}
	return db, nil
}

// MigrateDB runs GORM AutoMigrate for all models
func MigrateDB(ctx context.Context, db *gorm.DB) error {
	log := logging.S(ctx)
	log.Infow("Running DB automigration")
	if err := db.AutoMigrate(&Server{}, &IPAddress{}, &Billing{}, &EventLog{}); err != nil {
		log.Errorw("DB automigration failed", "error", err)
		return err
	}
	log.Infow("DB automigration complete")
	return nil
}
