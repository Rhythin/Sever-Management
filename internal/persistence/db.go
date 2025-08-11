package persistence

import (
	"context"
	"fmt"
	"net"

	"github.com/rhythin/sever-management/internal"
	"github.com/rhythin/sever-management/internal/logging"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDB initializes and migrates the database
func NewDB(cfg *internal.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	if err := MigrateDB(ctx, db, cfg); err != nil {
		return nil, err
	}
	return db, nil
}

// MigrateDB runs GORM AutoMigrate for all models
func MigrateDB(ctx context.Context, db *gorm.DB, cfg *internal.Config) error {
	log := logging.S(ctx)
	log.Infow("Running DB automigration")
	if err := db.AutoMigrate(&Server{}, &IPAddress{}, &Billing{}, &EventLog{}); err != nil {
		log.Errorw("DB automigration failed", "error", err)
		return err
	}
	log.Infow("DB automigration complete")
	// Seed IP addresses based on CIDR if none exist
	var count int64
	if err := db.WithContext(ctx).Model(&IPAddress{}).Count(&count).Error; err == nil && count == 0 {
		log.Infow("Seeding IP pool from CIDR")
		_, ipnet, perr := net.ParseCIDR(cfg.IPCIDR)
		if perr != nil {
			log.Errorw("Invalid IP CIDR in config; skipping seed", "cidr", cfg.IPCIDR, "error", perr)
		} else {
			// iterate IPs in CIDR, skip network and broadcast for IPv4, cap to a sane limit (e.g., 2048)
			seed := make([]IPAddress, 0, 1024)
			ip := ipnet.IP.Mask(ipnet.Mask)
			// advance to first usable
			for i := 0; i < 1<<uint(0); i++ { // no-op for clarity
			}
			max := 2048
			for ip := nextIP(ip); ipnet.Contains(ip) && len(seed) < max; ip = nextIP(ip) {
				seed = append(seed, IPAddress{Address: ip.String(), Allocated: false})
			}
			if len(seed) > 0 {
				if err := db.WithContext(ctx).Create(&seed).Error; err != nil {
					log.Errorw("Failed seeding IP pool", "error", err)
				} else {
					log.Infow("Seeded IP pool", "count", len(seed))
				}
			}
		}
	}
	return nil
}

// nextIP returns the next IPv4 address
func nextIP(ip net.IP) net.IP {
	nip := make(net.IP, len(ip))
	copy(nip, ip)
	for j := len(nip) - 1; j >= 0; j-- {
		nip[j]++
		if nip[j] != 0 {
			break
		}
	}
	return nip
}
