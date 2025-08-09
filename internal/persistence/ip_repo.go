package persistence

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IPRepo handles IP allocation and release

type IPRepo struct {
	db *gorm.DB
}

func NewIPRepo(db *gorm.DB) *IPRepo {
	return &IPRepo{db: db}
}

// AllocateIP atomically allocates an available IP and marks it as allocated
func (r *IPRepo) AllocateIP(ctx context.Context) (*IPAddress, error) {
	zap.S().Infow("IPRepo.AllocateIP called")
	var ip IPAddress
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("allocated = ?", false).First(&ip).Error; err != nil {
			zap.S().Warnw("IPRepo.AllocateIP no available IP", "error", err)
			return err
		}
		ip.Allocated = true
		return tx.Save(&ip).Error
	})
	if err != nil {
		zap.S().Errorw("IPRepo.AllocateIP failed", "error", err)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	zap.S().Infow("IPRepo.AllocateIP success", "ip", ip.Address, "ipID", ip.ID)
	return &ip, nil
}

// ReleaseIP marks an IP as unallocated
func (r *IPRepo) ReleaseIP(ctx context.Context, ipID uint) error {
	zap.S().Infow("IPRepo.ReleaseIP called", "ipID", ipID)
	err := r.db.WithContext(ctx).Model(&IPAddress{}).Where("id = ?", ipID).Update("allocated", false).Error
	if err != nil {
		zap.S().Errorw("IPRepo.ReleaseIP failed", "ipID", ipID, "error", err)
	}
	return err
}
