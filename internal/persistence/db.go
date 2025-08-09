package persistence

import (
	"fmt"

	"github.com/rhythin/sever-management/internal"
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
	// Auto-migrate all models
	if err := db.AutoMigrate(&Server{}, &IPAddress{}, &Billing{}, &EventLog{}); err != nil {
		return nil, err
	}
	return db, nil
}
