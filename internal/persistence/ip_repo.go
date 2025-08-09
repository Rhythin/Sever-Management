package persistence

import (
	"context"
	"errors"

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
	var ip IPAddress
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("allocated = ?", false).First(&ip).Error; err != nil {
			return err
		}
		ip.Allocated = true
		return tx.Save(&ip).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &ip, err
}

// ReleaseIP marks an IP as unallocated
func (r *IPRepo) ReleaseIP(ctx context.Context, ipID uint) error {
	return r.db.WithContext(ctx).Model(&IPAddress{}).Where("id = ?", ipID).Update("allocated", false).Error
}
