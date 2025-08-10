package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rhythin/sever-management/internal/domain"
	"github.com/rhythin/sever-management/internal/persistence"
)

// Mock repositories for testing
type mockServerRepo struct {
	servers map[string]*persistence.Server
	events  []*persistence.EventLog
}

func newMockServerRepo() *mockServerRepo {
	return &mockServerRepo{
		servers: make(map[string]*persistence.Server),
		events:  make([]*persistence.EventLog, 0),
	}
}

func (m *mockServerRepo) GetByID(ctx context.Context, id string) (*persistence.Server, error) {
	if server, exists := m.servers[id]; exists {
		return server, nil
	}
	return nil, errors.New("server not found")
}

func (m *mockServerRepo) Create(ctx context.Context, server *persistence.Server) error {
	if server.ID == "" {
		server.ID = "test-server-" + time.Now().Format("20060102150405")
	}
	m.servers[server.ID] = server
	return nil
}

func (m *mockServerRepo) UpdateState(ctx context.Context, id string, state string) error {
	if server, exists := m.servers[id]; exists {
		server.State = state
		return nil
	}
	return errors.New("server not found")
}

func (m *mockServerRepo) UpdateTimestamps(ctx context.Context, id string, started, stopped, terminated *time.Time) error {
	if server, exists := m.servers[id]; exists {
		server.StartedAt = started
		server.StoppedAt = stopped
		server.TerminatedAt = terminated
		return nil
	}
	return errors.New("server not found")
}

func (m *mockServerRepo) UpdateServer(ctx context.Context, id string, updates *persistence.Server) error {
	if server, exists := m.servers[id]; exists {
		if updates.State != "" {
			server.State = updates.State
		}
		if updates.StartedAt != nil {
			server.StartedAt = updates.StartedAt
		}
		return nil
	}
	return errors.New("server not found")
}

func (m *mockServerRepo) UpdateBilling(ctx context.Context, id string, accumulatedSeconds int64, totalCost float64) error {
	if server, exists := m.servers[id]; exists {
		if server.Billing == nil {
			server.Billing = &persistence.Billing{}
		}
		server.Billing.AccumulatedSeconds = accumulatedSeconds
		server.Billing.TotalCost = totalCost
		server.Billing.LastBilledAt = &time.Time{}
		*server.Billing.LastBilledAt = time.Now()
		return nil
	}
	return errors.New("server not found")
}

func (m *mockServerRepo) List(ctx context.Context, region, status, typ string, limit, offset int) ([]*persistence.Server, error) {
	var result []*persistence.Server
	for _, server := range m.servers {
		if region != "" && server.Region != region {
			continue
		}
		if status != "" && server.State != status {
			continue
		}
		if typ != "" && server.Type != typ {
			continue
		}
		result = append(result, server)
	}
	return result, nil
}

type mockIPRepo struct {
	ips map[uint]*persistence.IPAddress
}

func newMockIPRepo() *mockIPRepo {
	return &mockIPRepo{
		ips: make(map[uint]*persistence.IPAddress),
	}
}

func (m *mockIPRepo) AllocateIP(ctx context.Context) (*persistence.IPAddress, error) {
	// Simulate IP allocation
	ip := &persistence.IPAddress{
		ID:        uint(len(m.ips) + 1),
		Address:   "192.168.1.1",
		Allocated: true,
	}
	m.ips[ip.ID] = ip
	return ip, nil
}

func (m *mockIPRepo) ReleaseIP(ctx context.Context, id uint) error {
	if _, exists := m.ips[id]; exists {
		delete(m.ips, id)
		return nil
	}
	return errors.New("IP not found")
}

func (m *mockIPRepo) AssignIPToServer(ctx context.Context, ipID uint, serverID string) error {
	if ip, exists := m.ips[ipID]; exists {
		ip.ServerID = &serverID
		return nil
	}
	return errors.New("IP not found")
}

type mockEventRepo struct {
	events []*persistence.EventLog
}

func newMockEventRepo() *mockEventRepo {
	return &mockEventRepo{
		events: make([]*persistence.EventLog, 0),
	}
}

func (m *mockEventRepo) Append(ctx context.Context, event *persistence.EventLog) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockEventRepo) LastN(ctx context.Context, serverID string, n int) ([]persistence.EventLog, error) {
	var result []persistence.EventLog
	count := 0
	for i := len(m.events) - 1; i >= 0 && count < n; i-- {
		if m.events[i].ServerID == serverID {
			result = append([]persistence.EventLog{*m.events[i]}, result...)
			count++
		}
	}
	return result, nil
}

// Test helper function to create a service with mock repositories
func createTestService() (*ServerService, *mockServerRepo, *mockIPRepo, *mockEventRepo) {
	servers := newMockServerRepo()
	ips := newMockIPRepo()
	events := newMockEventRepo()

	// Create service using reflection or by temporarily modifying the service
	// For now, we'll test the individual functions directly
	return &ServerService{
		servers: servers,
		ips:     ips,
		events:  events,
	}, servers, ips, events
}

func TestNewServerService(t *testing.T) {
	servers := &persistence.ServerRepo{}
	ips := &persistence.IPRepo{}
	events := &persistence.EventRepo{}

	service := NewServerService(servers, ips, events)
	if service == nil {
		t.Fatal("NewServerService returned nil")
	}
}

func TestToDomainServer(t *testing.T) {
	now := time.Now()
	ps := &persistence.Server{
		ID:           "srv1",
		State:        string(domain.ServerRunning),
		StartedAt:    &now,
		StoppedAt:    nil,
		TerminatedAt: nil,
	}
	ds := toDomainServer(ps)
	if ds.ID != ps.ID {
		t.Errorf("ID = %s; want %s", ds.ID, ps.ID)
	}
	if ds.State != domain.ServerRunning {
		t.Errorf("State = %s; want %s", ds.State, domain.ServerRunning)
	}
	if ds.StartedAt == nil || !ds.StartedAt.Equal(now) {
		t.Errorf("StartedAt = %v; want %v", ds.StartedAt, now)
	}
	if ds.StoppedAt != nil {
		t.Errorf("StoppedAt = %v; want nil", ds.StoppedAt)
	}
	if ds.TerminatedAt != nil {
		t.Errorf("TerminatedAt = %v; want nil", ds.TerminatedAt)
	}
}

func TestServerServiceAction_Success(t *testing.T) {
	service, servers, _, _ := createTestService()

	// Create a test server in stopped state (can start from stopped)
	now := time.Now()
	server := &persistence.Server{
		ID:        "test-server",
		State:     string(domain.ServerStopped),
		StartedAt: &now,
	}
	servers.servers["test-server"] = server

	// Test start action
	err := service.Action(context.Background(), "test-server", domain.ActionStart)
	if err != nil {
		t.Errorf("Action failed: %v", err)
	}

	// Verify server state was updated
	updatedServer, _ := servers.GetByID(context.Background(), "test-server")
	if updatedServer.State != string(domain.ServerRunning) {
		t.Errorf("Expected state %s, got %s", string(domain.ServerRunning), updatedServer.State)
	}
}

func TestServerServiceAction_ServerNotFound(t *testing.T) {
	service, _, _, _ := createTestService()

	err := service.Action(context.Background(), "non-existent", domain.ActionStart)
	if err == nil {
		t.Error("Expected error for non-existent server")
	}
	if err.Error() != "server not found" {
		t.Errorf("Expected 'server not found' error, got: %v", err)
	}
}

func TestServerServiceAction_InvalidTransition(t *testing.T) {
	service, servers, _, _ := createTestService()

	// Create a terminated server (can't start)
	now := time.Now()
	server := &persistence.Server{
		ID:           "test-server",
		State:        string(domain.ServerTerminated),
		TerminatedAt: &now,
	}
	servers.servers["test-server"] = server

	// Try to start a terminated server
	err := service.Action(context.Background(), "test-server", domain.ActionStart)
	if err == nil {
		t.Error("Expected error for invalid transition")
	}
}

func TestServerServiceAction_UpdateStateFailure(t *testing.T) {
	servers := &mockServerRepoWithErrors{
		mockServerRepo:        newMockServerRepo(),
		shouldFailUpdateState: true,
	}
	ips := newMockIPRepo()
	events := newMockEventRepo()
	service := &ServerService{
		servers: servers,
		ips:     ips,
		events:  events,
	}

	// Create a test server in stopped state (can start from stopped)
	now := time.Now()
	server := &persistence.Server{
		ID:        "test-server",
		State:     string(domain.ServerStopped),
		StartedAt: &now,
	}
	servers.servers["test-server"] = server

	err := service.Action(context.Background(), "test-server", domain.ActionStart)
	if err == nil {
		t.Error("Expected error for update state failure")
	}
}

func TestServerServiceAction_UpdateTimestampsFailure(t *testing.T) {
	servers := &mockServerRepoWithErrors{
		mockServerRepo:             newMockServerRepo(),
		shouldFailUpdateTimestamps: true,
	}
	ips := newMockIPRepo()
	events := newMockEventRepo()
	service := &ServerService{
		servers: servers,
		ips:     ips,
		events:  events,
	}

	// Create a test server in stopped state (can start from stopped)
	now := time.Now()
	server := &persistence.Server{
		ID:        "test-server",
		State:     string(domain.ServerStopped),
		StartedAt: &now,
	}
	servers.servers["test-server"] = server

	err := service.Action(context.Background(), "test-server", domain.ActionStart)
	if err == nil {
		t.Error("Expected error for update timestamps failure")
	}
}

func TestServerServiceProvision_Success(t *testing.T) {
	service, servers, _, _ := createTestService()

	serverID, err := service.Provision(context.Background(), "us-west-1", "t2.micro")
	if err != nil {
		t.Errorf("Provision failed: %v", err)
	}
	if serverID == "" {
		t.Error("Expected server ID, got empty string")
	}

	// Verify server was created
	server, exists := servers.servers[serverID]
	if !exists {
		t.Error("Server was not created")
	}
	if server.Region != "us-west-1" {
		t.Errorf("Expected region us-west-1, got %s", server.Region)
	}
	if server.Type != "t2.micro" {
		t.Errorf("Expected type t2.micro, got %s", server.Type)
	}
}

func TestServerServiceProvision_NoAvailableIPs(t *testing.T) {
	servers := newMockServerRepo()
	ips := &mockIPRepoNoIPs{}
	events := newMockEventRepo()
	service := &ServerService{
		servers: servers,
		ips:     ips,
		events:  events,
	}

	_, err := service.Provision(context.Background(), "us-west-1", "t2.micro")
	if err == nil {
		t.Error("Expected error for no available IPs")
	}
	if err.Error() != "no available IPs" {
		t.Errorf("Expected 'no available IPs' error, got: %v", err)
	}
}

func TestServerServiceProvision_CreateServerFailure(t *testing.T) {
	servers := &mockServerRepoWithErrors{
		mockServerRepo:   newMockServerRepo(),
		shouldFailCreate: true,
	}
	ips := newMockIPRepo()
	events := newMockEventRepo()
	service := &ServerService{
		servers: servers,
		ips:     ips,
		events:  events,
	}

	_, err := service.Provision(context.Background(), "us-west-1", "t2.micro")
	if err == nil {
		t.Error("Expected error for server creation failure")
	}
}

func TestServerServiceProvision_AssignIPFailure(t *testing.T) {
	servers := newMockServerRepo()
	ips := &mockIPRepoWithErrors{shouldFailAssign: true}
	events := newMockEventRepo()
	service := &ServerService{
		servers: servers,
		ips:     ips,
		events:  events,
	}

	_, err := service.Provision(context.Background(), "us-west-1", "t2.micro")
	if err == nil {
		t.Error("Expected error for IP assignment failure")
	}
}

func TestServerServiceProvision_UpdateStateFailure(t *testing.T) {
	servers := &mockServerRepoWithErrors{
		mockServerRepo:         newMockServerRepo(),
		shouldFailUpdateServer: true,
	}
	_ = newMockIPRepo()
	_ = newMockEventRepo()
	service := &ServerService{
		servers: servers,
		ips:     newMockIPRepo(),
		events:  newMockEventRepo(),
	}

	serverID, err := service.Provision(context.Background(), "us-west-1", "t2.micro")
	if err == nil {
		t.Error("Expected error for state update failure")
	}
	// Server should still be created even if state update fails
	if serverID == "" {
		t.Error("Expected server ID even with state update failure")
	}
}

func TestServerServiceProvision_EventLoggingFailure(t *testing.T) {
	servers := newMockServerRepo()
	ips := newMockIPRepo()
	events := &mockEventRepoWithErrors{shouldFailAppend: true}
	service := &ServerService{
		servers: servers,
		ips:     ips,
		events:  events,
	}

	serverID, err := service.Provision(context.Background(), "us-west-1", "t2.micro")
	if err == nil {
		t.Error("Expected error for event logging failure")
	}
	// Server should still be created even if event logging fails
	if serverID == "" {
		t.Error("Expected server ID even with event logging failure")
	}
}

func TestServerServiceGetEvents(t *testing.T) {
	service, _, _, events := createTestService()

	// Add some test events
	testEvent := &persistence.EventLog{
		ServerID:  "test-server",
		Timestamp: time.Now(),
		Type:      "test",
		Message:   "test message",
	}
	events.events = append(events.events, testEvent)

	result, err := service.GetEvents(context.Background(), "test-server", 10)
	if err != nil {
		t.Errorf("GetEvents failed: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("Expected 1 event, got %d", len(result))
	}
	if result[0].Message != "test message" {
		t.Errorf("Expected message 'test message', got %s", result[0].Message)
	}
}

// Mock repositories with error conditions
type mockServerRepoWithErrors struct {
	*mockServerRepo
	shouldFailCreate           bool
	shouldFailUpdateState      bool
	shouldFailUpdateTimestamps bool
	shouldFailUpdateServer     bool
}

func (m *mockServerRepoWithErrors) Create(ctx context.Context, server *persistence.Server) error {
	if m.shouldFailCreate {
		return errors.New("create failed")
	}
	return m.mockServerRepo.Create(ctx, server)
}

func (m *mockServerRepoWithErrors) UpdateState(ctx context.Context, id string, state string) error {
	if m.shouldFailUpdateState {
		return errors.New("update state failed")
	}
	return m.mockServerRepo.UpdateState(ctx, id, state)
}

func (m *mockServerRepoWithErrors) UpdateTimestamps(ctx context.Context, id string, started, stopped, terminated *time.Time) error {
	if m.shouldFailUpdateTimestamps {
		return errors.New("update timestamps failed")
	}
	return m.mockServerRepo.UpdateTimestamps(ctx, id, started, stopped, terminated)
}

func (m *mockServerRepoWithErrors) UpdateServer(ctx context.Context, id string, updates *persistence.Server) error {
	if m.shouldFailUpdateServer {
		return errors.New("update server failed")
	}
	return m.mockServerRepo.UpdateServer(ctx, id, updates)
}

type mockIPRepoNoIPs struct {
	*mockIPRepo
}

func (m *mockIPRepoNoIPs) AllocateIP(ctx context.Context) (*persistence.IPAddress, error) {
	return nil, nil // Return nil to simulate no available IPs
}

type mockIPRepoWithErrors struct {
	*mockIPRepo
	shouldFailAssign bool
}

func (m *mockIPRepoWithErrors) AllocateIP(ctx context.Context) (*persistence.IPAddress, error) {
	// Simulate IP allocation
	ip := &persistence.IPAddress{
		ID:        uint(len(m.ips) + 1),
		Address:   "192.168.1.1",
		Allocated: true,
	}
	m.ips[ip.ID] = ip
	return ip, nil
}

func (m *mockIPRepoWithErrors) AssignIPToServer(ctx context.Context, ipID uint, serverID string) error {
	if m.shouldFailAssign {
		return errors.New("assign IP failed")
	}
	return m.mockIPRepo.AssignIPToServer(ctx, ipID, serverID)
}

type mockEventRepoWithErrors struct {
	*mockEventRepo
	shouldFailAppend bool
}

func (m *mockEventRepoWithErrors) Append(ctx context.Context, event *persistence.EventLog) error {
	if m.shouldFailAppend {
		return errors.New("append failed")
	}
	return m.mockEventRepo.Append(ctx, event)
}
