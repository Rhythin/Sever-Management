package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&Server{}, &IPAddress{}, &Billing{}, &EventLog{})
	require.NoError(t, err)

	return db
}

func TestServerRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepo(db)
	ctx := context.Background()

	server := &Server{
		ID:     "test-123",
		Type:   "t2.micro",
		Region: "us-west-1",
		State:  "stopped",
	}

	err := repo.Create(ctx, server)
	assert.NoError(t, err)
	assert.NotZero(t, server.ID)

	// Verify it was created
	var found Server
	err = db.First(&found, "id = ?", server.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "t2.micro", found.Type)
}

func TestServerRepo_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepo(db)
	ctx := context.Background()

	// Create a test server
	server := &Server{
		ID:     "test-123",
		Type:   "t2.micro",
		Region: "us-west-1",
		State:  "stopped",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// Get by ID
	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.Equal(t, "t2.micro", found.Type)

	// Test non-existent ID
	notFound, err := repo.GetByID(ctx, "non-existent")
	assert.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestServerRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepo(db)
	ctx := context.Background()

	// Create test servers
	servers := []*Server{
		{ID: "1", Type: "t2.micro", Region: "us-west-1", State: "running"},
		{ID: "2", Type: "t2.small", Region: "us-west-2", State: "stopped"},
		{ID: "3", Type: "t2.micro", Region: "us-west-1", State: "running"},
	}

	for _, s := range servers {
		err := repo.Create(ctx, s)
		require.NoError(t, err)
	}

	// Test listing all servers
	all, err := repo.List(ctx, "", "", "", 100, 0)
	assert.NoError(t, err)
	assert.Len(t, all, 3)

	// Test filtering by region
	west1, err := repo.List(ctx, "us-west-1", "", "", 100, 0)
	assert.NoError(t, err)
	assert.Len(t, west1, 2)

	// Test filtering by state
	running, err := repo.List(ctx, "", "running", "", 100, 0)
	assert.NoError(t, err)
	assert.Len(t, running, 2)

	// Test filtering by type
	micro, err := repo.List(ctx, "", "", "t2.micro", 100, 0)
	assert.NoError(t, err)
	assert.Len(t, micro, 2)

	// Test pagination
	limited, err := repo.List(ctx, "", "", "", 2, 0)
	assert.NoError(t, err)
	assert.Len(t, limited, 2)

	offset, err := repo.List(ctx, "", "", "", 1, 1)
	assert.NoError(t, err)
	assert.Len(t, offset, 1)
}

func TestServerRepo_UpdateState(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepo(db)
	ctx := context.Background()

	// Create a test server
	server := &Server{
		ID:     "test-123",
		Type:   "t2.micro",
		Region: "us-west-1",
		State:  "stopped",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// Update state
	err = repo.UpdateState(ctx, server.ID, "running")
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.Equal(t, "running", found.State)
}

func TestServerRepo_UpdateTimestamps(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepo(db)
	ctx := context.Background()

	// Create a test server
	server := &Server{
		ID:     "test-123",
		Type:   "t2.micro",
		Region: "us-west-1",
		State:  "stopped",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	now := time.Now()
	startedAt := &now
	stoppedAt := &now

	// Update timestamps
	err = repo.UpdateTimestamps(ctx, server.ID, startedAt, stoppedAt, nil)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.Equal(t, startedAt.Unix(), found.StartedAt.Unix())
	assert.Equal(t, stoppedAt.Unix(), found.StoppedAt.Unix())
}

func TestServerRepo_UpdateBilling(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServerRepo(db)
	ctx := context.Background()

	// Create a test server
	server := &Server{
		ID:     "test-123",
		Type:   "t2.micro",
		Region: "us-west-1",
		State:  "stopped",
	}
	err := repo.Create(ctx, server)
	require.NoError(t, err)

	// Create a billing record first
	billing := &Billing{
		ServerID:           server.ID,
		AccumulatedSeconds: 0,
		TotalCost:          0.0,
	}
	err = db.Create(billing).Error
	require.NoError(t, err)

	// Update billing
	err = repo.UpdateBilling(ctx, server.ID, 3600, 0.10)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.GetByID(ctx, server.ID)
	assert.NoError(t, err)
	assert.Equal(t, int64(3600), found.Billing.AccumulatedSeconds)
	assert.Equal(t, 0.10, found.Billing.TotalCost)
}
