package persistence

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"gorm.io/gorm"
)

func Test_ipRepo_AllocateIP(t *testing.T) {
	type fields struct {
		db *gorm.DB
		mu sync.Mutex
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *IPAddress
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ipRepo{
				db: tt.fields.db,
				mu: tt.fields.mu,
			}
			got, err := r.AllocateIP(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ipRepo.AllocateIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ipRepo.AllocateIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ipRepo_ReleaseIP(t *testing.T) {
	type fields struct {
		db *gorm.DB
		mu sync.Mutex
	}
	type args struct {
		ctx  context.Context
		ipID uint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ipRepo{
				db: tt.fields.db,
				mu: tt.fields.mu,
			}
			if err := r.ReleaseIP(tt.args.ctx, tt.args.ipID); (err != nil) != tt.wantErr {
				t.Errorf("ipRepo.ReleaseIP() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_ipRepo_AssignIPToServer(t *testing.T) {
	type fields struct {
		db *gorm.DB
		mu sync.Mutex
	}
	type args struct {
		ctx      context.Context
		ipID     uint
		serverID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ipRepo{
				db: tt.fields.db,
				mu: tt.fields.mu,
			}
			if err := r.AssignIPToServer(tt.args.ctx, tt.args.ipID, tt.args.serverID); (err != nil) != tt.wantErr {
				t.Errorf("ipRepo.AssignIPToServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
