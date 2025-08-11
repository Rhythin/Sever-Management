package persistence

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gorm.io/gorm"
)

func Test_serverRepo_Create(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx context.Context
		s   *Server
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
			r := &serverRepo{
				db: tt.fields.db,
			}
			if err := r.Create(tt.args.ctx, tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("serverRepo.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_serverRepo_GetByID(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Server
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &serverRepo{
				db: tt.fields.db,
			}
			got, err := r.GetByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("serverRepo.GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serverRepo.GetByID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serverRepo_UpdateState(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx      context.Context
		id       string
		newState string
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
			r := &serverRepo{
				db: tt.fields.db,
			}
			if err := r.UpdateState(tt.args.ctx, tt.args.id, tt.args.newState); (err != nil) != tt.wantErr {
				t.Errorf("serverRepo.UpdateState() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_serverRepo_List(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx    context.Context
		region string
		status string
		typ    string
		limit  int
		offset int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Server
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &serverRepo{
				db: tt.fields.db,
			}
			got, err := r.List(tt.args.ctx, tt.args.region, tt.args.status, tt.args.typ, tt.args.limit, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("serverRepo.List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serverRepo.List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serverRepo_UpdateTimestamps(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx        context.Context
		id         string
		started    *time.Time
		stopped    *time.Time
		terminated *time.Time
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
			r := &serverRepo{
				db: tt.fields.db,
			}
			if err := r.UpdateTimestamps(tt.args.ctx, tt.args.id, tt.args.started, tt.args.stopped, tt.args.terminated); (err != nil) != tt.wantErr {
				t.Errorf("serverRepo.UpdateTimestamps() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_serverRepo_UpdateBilling(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx                context.Context
		id                 string
		accumulatedSeconds int64
		totalCost          float64
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
			r := &serverRepo{
				db: tt.fields.db,
			}
			if err := r.UpdateBilling(tt.args.ctx, tt.args.id, tt.args.accumulatedSeconds, tt.args.totalCost); (err != nil) != tt.wantErr {
				t.Errorf("serverRepo.UpdateBilling() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_serverRepo_UpdateServer(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx    context.Context
		id     string
		server *Server
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
			r := &serverRepo{
				db: tt.fields.db,
			}
			if err := r.UpdateServer(tt.args.ctx, tt.args.id, tt.args.server); (err != nil) != tt.wantErr {
				t.Errorf("serverRepo.UpdateServer() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
