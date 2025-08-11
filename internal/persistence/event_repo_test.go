package persistence

import (
	"context"
	"reflect"
	"testing"

	"gorm.io/gorm"
)

func Test_eventRepo_Append(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx   context.Context
		event *EventLog
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
			r := &eventRepo{
				db: tt.fields.db,
			}
			if err := r.Append(tt.args.ctx, tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("eventRepo.Append() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_eventRepo_LastN(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx      context.Context
		serverID string
		n        int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []EventLog
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &eventRepo{
				db: tt.fields.db,
			}
			got, err := r.LastN(tt.args.ctx, tt.args.serverID, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("eventRepo.LastN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("eventRepo.LastN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventRepo_GetEvents(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx      context.Context
		serverID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []EventLog
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &eventRepo{
				db: tt.fields.db,
			}
			got, err := r.GetEvents(tt.args.ctx, tt.args.serverID)
			if (err != nil) != tt.wantErr {
				t.Errorf("eventRepo.GetEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("eventRepo.GetEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_eventRepo_AddEvent(t *testing.T) {
	type fields struct {
		db *gorm.DB
	}
	type args struct {
		ctx   context.Context
		event *EventLog
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
			r := &eventRepo{
				db: tt.fields.db,
			}
			if err := r.AddEvent(tt.args.ctx, tt.args.event); (err != nil) != tt.wantErr {
				t.Errorf("eventRepo.AddEvent() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
