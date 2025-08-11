package service

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/persistence"
	mockPersistence "github.com/rhythin/sever-management/internal/persistence/mocks"
	"github.com/stretchr/testify/mock"
)

func Test_serverService_Action(t *testing.T) {

	validServer := &persistence.Server{
		ID:           "2",
		State:        "running",
		StartedAt:    &time.Time{},
		StoppedAt:    &time.Time{},
		TerminatedAt: &time.Time{},
		Events:       []*persistence.EventLog{},
	}

	stoppedServer := &persistence.Server{
		ID:           "5",
		State:        "stopped",
		StartedAt:    &time.Time{},
		StoppedAt:    &time.Time{},
		TerminatedAt: &time.Time{},
		Events:       []*persistence.EventLog{},
	}

	mockServerRepo := &mockPersistence.ServerRepoInterface{}
	mockServerRepo.On("GetByID", mock.Anything, "1").Return(nil, nil)
	mockServerRepo.On("GetByID", mock.Anything, "2").Return(validServer, nil)
	mockServerRepo.On("GetByID", mock.Anything, "3").Return(validServer, nil)
	mockServerRepo.On("GetByID", mock.Anything, "4").Return(validServer, nil)
	mockServerRepo.On("GetByID", mock.Anything, "5").Return(stoppedServer, nil)
	mockServerRepo.On("UpdateState", mock.Anything, "2", "stopped").Return(errors.New("update state failed")).Once()
	mockServerRepo.On("UpdateState", mock.Anything, "3", "stopped").Return(nil)
	mockServerRepo.On("UpdateState", mock.Anything, "4", "stopped").Return(nil)
	mockServerRepo.On("UpdateState", mock.Anything, "4", "terminated").Return(nil)
	mockServerRepo.On("UpdateState", mock.Anything, "5", "running").Return(nil)
	mockServerRepo.On("UpdateState", mock.Anything, "5", "terminated").Return(nil)
	mockServerRepo.On("UpdateTimestamps", mock.Anything, "3", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("update timestamps failed")).Once()
	mockServerRepo.On("UpdateTimestamps", mock.Anything, "4", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockServerRepo.On("UpdateTimestamps", mock.Anything, "5", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockIPRepo := &mockPersistence.IPRepoInterface{}

	mockEventRepo := &mockPersistence.EventRepoInterface{}
	mockEventRepo.On("Append", mock.Anything, mock.Anything).Return(nil)

	type fields struct {
		servers persistence.ServerRepo
		ips     persistence.IPRepo
		events  persistence.EventRepo
	}
	type args struct {
		ctx    context.Context
		id     string
		action domain.ServerAction
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "getByID error persistence layer",
			args: args{
				ctx:    context.Background(),
				id:     "1",
				action: domain.ActionStart,
			},
			wantErr: true,
			fields: fields{
				servers: mockServerRepo,
				ips:     mockIPRepo,
				events:  mockEventRepo,
			},
		},
		{
			name: "invalid state transition",
			args: args{
				ctx:    context.Background(),
				id:     "2",
				action: domain.ServerAction("invalid-action"),
			},
			wantErr: true,
			fields: fields{
				servers: mockServerRepo,
				ips:     mockIPRepo,
				events:  mockEventRepo,
			},
		},
		{
			name: "update state error persistence layer",
			args: args{
				ctx:    context.Background(),
				id:     "2",
				action: domain.ActionStop,
			},
			wantErr: true,
			fields: fields{
				servers: mockServerRepo,
				ips:     mockIPRepo,
				events:  mockEventRepo,
			},
		},
		{
			name: "update timestamps error persistence layer",
			args: args{
				ctx:    context.Background(),
				id:     "3",
				action: domain.ActionStop,
			},
			wantErr: true,
			fields: fields{
				servers: mockServerRepo,
				ips:     mockIPRepo,
				events:  mockEventRepo,
			},
		},
		{
			name: "update state success start",
			args: args{
				ctx:    context.Background(),
				id:     "5",
				action: domain.ActionStart,
			},
			wantErr: false,
			fields: fields{
				servers: mockServerRepo,
				ips:     mockIPRepo,
				events:  mockEventRepo,
			},
		},
		{
			name: "update state success stop",
			args: args{
				ctx:    context.Background(),
				id:     "4",
				action: domain.ActionStop,
			},
			wantErr: false,
			fields: fields{
				servers: mockServerRepo,
				ips:     mockIPRepo,
				events:  mockEventRepo,
			},
		},
		{
			name: "update state success terminate",
			args: args{
				ctx:    context.Background(),
				id:     "4",
				action: domain.ActionTerminate,
			},
			wantErr: false,
			fields: fields{
				servers: mockServerRepo,
				ips:     mockIPRepo,
				events:  mockEventRepo,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serverService{
				servers: tt.fields.servers,
				ips:     tt.fields.ips,
				events:  tt.fields.events,
			}
			if err := s.Action(tt.args.ctx, tt.args.id, tt.args.action); (err != nil) != tt.wantErr {
				t.Errorf("serverService.Action() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_serverService_Provision(t *testing.T) {
	type fields struct {
		servers persistence.ServerRepo
		ips     persistence.IPRepo
		events  persistence.EventRepo
	}
	type args struct {
		ctx    context.Context
		region string
		typ    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Allocate IP error",
			fields: fields{
				servers: &mockPersistence.ServerRepoInterface{},
				ips: func() *mockPersistence.IPRepoInterface {
					mock := &mockPersistence.IPRepoInterface{}
					mock.On("AllocateIP", context.Background()).Return(nil, errors.New("allocation error"))
					return mock
				}(),
				events: &mockPersistence.EventRepoInterface{},
			},
			args: args{
				ctx:    context.Background(),
				region: "us-west-1",
				typ:    "t2.micro",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "No available IPs",
			fields: fields{
				servers: &mockPersistence.ServerRepoInterface{},
				ips: func() *mockPersistence.IPRepoInterface {
					mock := &mockPersistence.IPRepoInterface{}
					mock.On("AllocateIP", context.Background()).Return(nil, nil)
					return mock
				}(),
				events: &mockPersistence.EventRepoInterface{},
			},
			args: args{
				ctx:    context.Background(),
				region: "us-west-1",
				typ:    "t2.micro",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Server creation failure",
			fields: fields{
				servers: func() *mockPersistence.ServerRepoInterface {
					mockServerRepo := &mockPersistence.ServerRepoInterface{}
					mockServerRepo.On("Create", context.Background(), mock.Anything).Return(errors.New("create error"))
					return mockServerRepo
				}(),
				ips: func() *mockPersistence.IPRepoInterface {
					mockIPRepo := &mockPersistence.IPRepoInterface{}
					mockIPRepo.On("AllocateIP", context.Background()).Return(&persistence.IPAddress{ID: 1, Address: "192.168.1.1"}, nil)
					mockIPRepo.On("ReleaseIP", context.Background(), uint(1)).Return(nil)
					return mockIPRepo
				}(),
				events: &mockPersistence.EventRepoInterface{},
			},
			args: args{
				ctx:    context.Background(),
				region: "us-west-1",
				typ:    "t2.micro",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "IP assignment failure",
			fields: fields{
				servers: func() *mockPersistence.ServerRepoInterface {
					mockServerRepo := &mockPersistence.ServerRepoInterface{}
					mockServerRepo.On("Create", context.Background(), mock.Anything).Run(func(args mock.Arguments) {
						server := args.Get(1).(*persistence.Server)
						server.ID = "test-server"
					}).Return(nil)
					mockServerRepo.On("Delete", context.Background(), "test-server").Return(nil)
					return mockServerRepo
				}(),
				ips: func() *mockPersistence.IPRepoInterface {
					mockIPRepo := &mockPersistence.IPRepoInterface{}
					mockIPRepo.On("AllocateIP", context.Background()).Return(&persistence.IPAddress{ID: 1, Address: "192.168.1.1"}, nil)
					mockIPRepo.On("AssignIPToServer", context.Background(), uint(1), "test-server").Return(errors.New("assign error"))
					mockIPRepo.On("ReleaseIP", context.Background(), uint(1)).Return(nil)
					return mockIPRepo
				}(),
				events: &mockPersistence.EventRepoInterface{},
			},
			args: args{
				ctx:    context.Background(),
				region: "us-west-1",
				typ:    "t2.micro",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Successful provisioning",
			fields: fields{
				servers: func() *mockPersistence.ServerRepoInterface {
					mockServerRepo := &mockPersistence.ServerRepoInterface{}
					mockServerRepo.On("Create", context.Background(), mock.Anything).Run(func(args mock.Arguments) {
						s := args.Get(1).(*persistence.Server)
						s.ID = "test-server"
					}).Return(nil)
					mockServerRepo.On("UpdateServer", context.Background(), "test-server", mock.Anything).Return(nil)
					return mockServerRepo
				}(),
				ips: func() *mockPersistence.IPRepoInterface {
					mockIPRepo := &mockPersistence.IPRepoInterface{}
					mockIPRepo.On("AllocateIP", context.Background()).Return(&persistence.IPAddress{
						ID:      1,
						Address: "192.168.1.1",
					}, nil)
					mockIPRepo.On("AssignIPToServer", context.Background(), uint(1), "test-server").Return(nil)
					return mockIPRepo
				}(),
				events: func() *mockPersistence.EventRepoInterface {
					mockEventRepo := &mockPersistence.EventRepoInterface{}
					mockEventRepo.On("Append", context.Background(), mock.Anything).Return(nil).Twice()
					return mockEventRepo
				}(),
			},
			args: args{
				ctx:    context.Background(),
				region: "us-west-1",
				typ:    "t2.micro",
			},
			want:    "test-server",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serverService{
				servers: tt.fields.servers,
				ips:     tt.fields.ips,
				events:  tt.fields.events,
			}
			got, err := s.Provision(tt.args.ctx, tt.args.region, tt.args.typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("serverService.Provision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("serverService.Provision() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toDomainServer(t *testing.T) {
	request := &persistence.Server{
		ID:           "1",
		State:        "running",
		StartedAt:    &time.Time{},
		StoppedAt:    &time.Time{},
		TerminatedAt: &time.Time{},
	}

	expected := &domain.Server{
		ID:           "1",
		State:        domain.ServerState("running"),
		StartedAt:    &time.Time{},
		StoppedAt:    &time.Time{},
		TerminatedAt: &time.Time{},
		Log:          domain.NewEventRingBuffer(100),
	}

	type args struct {
		s *persistence.Server
	}
	tests := []struct {
		name string
		args args
		want *domain.Server
	}{
		{
			name: "toDomainServer success",
			args: args{
				s: request,
			},
			want: expected,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toDomainServer(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toDomainServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serverService_GetEvents(t *testing.T) {
	mockEventRepo := &mockPersistence.EventRepoInterface{}

	mockEventRepo.On("LastN", mock.Anything, "id", 10).Return([]persistence.EventLog{}, nil)

	type fields struct {
		servers persistence.ServerRepo
		ips     persistence.IPRepo
		events  persistence.EventRepo
	}
	type args struct {
		ctx context.Context
		id  string
		n   int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []persistence.EventLog
		wantErr bool
	}{
		{
			name: "GetEvents success",
			args: args{
				ctx: context.Background(),
				id:  "id",
				n:   10,
			},
			want:    []persistence.EventLog{},
			wantErr: false,
			fields: fields{
				events: mockEventRepo,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serverService{
				servers: tt.fields.servers,
				ips:     tt.fields.ips,
				events:  tt.fields.events,
			}
			got, err := s.GetEvents(tt.args.ctx, tt.args.id, tt.args.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("serverService.GetEvents() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serverService.GetEvents() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serverService_ListServers(t *testing.T) {
	mockServerRepo := &mockPersistence.ServerRepoInterface{}

	mockServerRepo.On("List", mock.Anything, "region", "status", "type", 20, 0).Return([]*persistence.Server{}, nil)

	type fields struct {
		servers persistence.ServerRepo
		ips     persistence.IPRepo
		events  persistence.EventRepo
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
		want    []*persistence.Server
		wantErr bool
	}{
		{
			name: "ListServers success",
			args: args{
				ctx:    context.Background(),
				region: "region",
				status: "status",
				typ:    "type",
				limit:  20,
				offset: 0,
			},
			want:    []*persistence.Server{},
			wantErr: false,
			fields: fields{
				servers: mockServerRepo,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serverService{
				servers: tt.fields.servers,
				ips:     tt.fields.ips,
				events:  tt.fields.events,
			}
			got, err := s.ListServers(tt.args.ctx, tt.args.region, tt.args.status, tt.args.typ, tt.args.limit, tt.args.offset)
			if (err != nil) != tt.wantErr {
				t.Errorf("serverService.ListServers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serverService.ListServers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_serverService_GetServerByID(t *testing.T) {

	mockServerRepo := &mockPersistence.ServerRepoInterface{}

	mockServerRepo.On("GetByID", mock.Anything, "1").Return(&persistence.Server{}, nil)

	type fields struct {
		servers persistence.ServerRepo
		ips     persistence.IPRepo
		events  persistence.EventRepo
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *persistence.Server
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "GetServerByID success",
			args: args{
				ctx: context.Background(),
				id:  "1",
			},
			want: &persistence.Server{},
			fields: fields{
				servers: mockServerRepo,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &serverService{
				servers: tt.fields.servers,
				ips:     tt.fields.ips,
				events:  tt.fields.events,
			}
			got, err := s.GetServerByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("serverService.GetServerByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("serverService.GetServerByID() = %v, want %v", got, tt.want)
			}
		})
	}
}
