package persistence

import (
	"context"
	"errors"
	"sync"

	"github.com/rhythin/sever-management/internal/logging"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IPRepo handles IP allocation and release

type ipRepo struct {
	db *gorm.DB
	mu sync.Mutex // serialize allocation attempts for extra safety
}

func NewIPRepo(db *gorm.DB) IPRepo {
	return &ipRepo{db: db}
}

// AllocateIP atomically allocates an available IP and marks it as allocated
// Uses GORM transaction with row-level locking and a mutex for extra thread safety
func (r *ipRepo) AllocateIP(ctx context.Context) (*IPAddress, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	log := logging.S(ctx)
	log.Infow("IPRepo.AllocateIP called")
	var ip IPAddress
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("allocated = ?", false).First(&ip).Error; err != nil {
			// Treat no rows as a normal condition (no available IPs)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Warnw("IPRepo.AllocateIP no available IP")
			} else {
				log.Errorw("IPRepo.AllocateIP query failed", "error", err)
			}
			return err
		}
		ip.Allocated = true
		return tx.Save(&ip).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		log.Errorw("IPRepo.AllocateIP failed", "error", err)
		return nil, err
	}
	log.Infow("IPRepo.AllocateIP success", "ip", ip.Address, "ipID", ip.ID)
	return &ip, nil
}

func (r *ipRepo) ReleaseIP(ctx context.Context, ipID uint) error {
	log := logging.S(ctx)
	log.Infow("IPRepo.ReleaseIP called", "ipID", ipID)
	err := r.db.WithContext(ctx).Model(&IPAddress{}).Where("id = ?", ipID).Updates(map[string]interface{}{
		"allocated": false,
		"server_id": nil,
	}).Error
	if err != nil {
		log.Errorw("IPRepo.ReleaseIP failed", "ipID", ipID, "error", err)
	}
	return err
}

// AssignIPToServer links an IP to a server record
func (r *ipRepo) AssignIPToServer(ctx context.Context, ipID uint, serverID string) error {
	log := logging.S(ctx)
	log.Infow("IPRepo.AssignIPToServer called", "ipID", ipID, "serverID", serverID)

	// Check if IP exists first
	var ip IPAddress
	if err := r.db.WithContext(ctx).First(&ip, ipID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warnw("IPRepo.AssignIPToServer IP not found", "ipID", ipID)
			return errors.New("IP not found")
		}
		log.Errorw("IPRepo.AssignIPToServer query failed", "ipID", ipID, "error", err)
		return err
	}

	err := r.db.WithContext(ctx).Model(&IPAddress{}).Where("id = ?", ipID).Update("server_id", serverID).Error
	if err != nil {
		log.Errorw("IPRepo.AssignIPToServer failed", "ipID", ipID, "serverID", serverID, "error", err)
	}
	return err
}
