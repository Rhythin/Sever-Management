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

func setupTestEventDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Manually create the table with the correct name
	err = db.Exec(`CREATE TABLE IF NOT EXISTS event_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		server_id TEXT,
		timestamp DATETIME,
		type TEXT,
		message TEXT
	)`).Error
	require.NoError(t, err)

	return db
}

func TestEventRepo_AddEvent(t *testing.T) {
	db := setupTestEventDB(t)
	repo := NewEventRepo(db)
	ctx := context.Background()

	event := &EventLog{
		ServerID:  "server-123",
		Timestamp: time.Now(),
		Type:      "started",
		Message:   "Server started successfully",
	}

	err := repo.AddEvent(ctx, event)
	assert.NoError(t, err)
	assert.NotZero(t, event.ID)

	// Verify it was created
	var found EventLog
	err = db.First(&found, event.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "server-123", found.ServerID)
	assert.Equal(t, "started", found.Type)
	assert.Equal(t, "Server started successfully", found.Message)
}

func TestEventRepo_GetEvents(t *testing.T) {
	db := setupTestEventDB(t)
	repo := NewEventRepo(db)
	ctx := context.Background()

	// Create test events
	now := time.Now()
	events := []*EventLog{
		{
			ServerID:  "server-123",
			Timestamp: now.Add(-2 * time.Hour),
			Type:      "provisioned",
			Message:   "Server provisioned",
		},
		{
			ServerID:  "server-123",
			Timestamp: now.Add(-1 * time.Hour),
			Type:      "started",
			Message:   "Server started",
		},
		{
			ServerID:  "server-123",
			Timestamp: now,
			Type:      "stopped",
			Message:   "Server stopped",
		},
	}

	for _, event := range events {
		err := repo.AddEvent(ctx, event)
		require.NoError(t, err)
	}

	// Get events for the server
	retrieved, err := repo.GetEvents(ctx, "server-123")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// Events should be ordered by timestamp (newest first due to GORM default)
	assert.Equal(t, "stopped", retrieved[0].Type)
	assert.Equal(t, "started", retrieved[1].Type)
	assert.Equal(t, "provisioned", retrieved[2].Type)
}

func TestEventRepo_GetEvents_NonExistentServer(t *testing.T) {
	db := setupTestEventDB(t)
	repo := NewEventRepo(db)
	ctx := context.Background()

	// Try to get events for non-existent server
	events, err := repo.GetEvents(ctx, "non-existent")
	assert.NoError(t, err)
	assert.Len(t, events, 0)
}

func TestEventRepo_GetEvents_EmptyServer(t *testing.T) {
	db := setupTestEventDB(t)
	repo := NewEventRepo(db)
	ctx := context.Background()

	// Get events for server with no events
	events, err := repo.GetEvents(ctx, "server-123")
	assert.NoError(t, err)
	assert.Len(t, events, 0)
}

func TestEventRepo_AddEvent_MultipleServers(t *testing.T) {
	db := setupTestEventDB(t)
	repo := NewEventRepo(db)
	ctx := context.Background()

	// Create events for multiple servers
	event1 := &EventLog{
		ServerID:  "server-1",
		Timestamp: time.Now(),
		Type:      "started",
		Message:   "Server 1 started",
	}

	event2 := &EventLog{
		ServerID:  "server-2",
		Timestamp: time.Now(),
		Type:      "started",
		Message:   "Server 2 started",
	}

	err := repo.AddEvent(ctx, event1)
	assert.NoError(t, err)

	err = repo.AddEvent(ctx, event2)
	assert.NoError(t, err)

	// Get events for each server
	events1, err := repo.GetEvents(ctx, "server-1")
	assert.NoError(t, err)
	assert.Len(t, events1, 1)
	assert.Equal(t, "server-1", events1[0].ServerID)

	events2, err := repo.GetEvents(ctx, "server-2")
	assert.NoError(t, err)
	assert.Len(t, events2, 1)
	assert.Equal(t, "server-2", events2[0].ServerID)
}

func TestEventRepo_AddEvent_Concurrent(t *testing.T) {
	db := setupTestEventDB(t)
	repo := NewEventRepo(db)
	ctx := context.Background()

	// Test concurrent event creation
	results := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			event := &EventLog{
				ServerID:  "server-123",
				Timestamp: time.Now(),
				Type:      "test",
				Message:   "Test event",
			}
			err := repo.AddEvent(ctx, event)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < 10; i++ {
		err := <-results
		assert.NoError(t, err)
	}

	// Verify all events were created
	events, err := repo.GetEvents(ctx, "server-123")
	assert.NoError(t, err)
	assert.Len(t, events, 10)
}

func TestEventRepo_EventOrdering(t *testing.T) {
	db := setupTestEventDB(t)
	repo := NewEventRepo(db)
	ctx := context.Background()

	// Create events with specific timestamps
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	events := []*EventLog{
		{
			ServerID:  "server-123",
			Timestamp: baseTime.Add(1 * time.Second),
			Type:      "second",
			Message:   "Second event",
		},
		{
			ServerID:  "server-123",
			Timestamp: baseTime,
			Type:      "first",
			Message:   "First event",
		},
		{
			ServerID:  "server-123",
			Timestamp: baseTime.Add(2 * time.Second),
			Type:      "third",
			Message:   "Third event",
		},
	}

	// Add events in random order
	repo.AddEvent(ctx, events[1]) // First
	repo.AddEvent(ctx, events[2]) // Third
	repo.AddEvent(ctx, events[0]) // Second

	// Get events and verify ordering
	retrieved, err := repo.GetEvents(ctx, "server-123")
	assert.NoError(t, err)
	assert.Len(t, retrieved, 3)

	// Should be ordered by timestamp (newest first due to GORM default)
	assert.Equal(t, "third", retrieved[0].Type)
	assert.Equal(t, "second", retrieved[1].Type)
	assert.Equal(t, "first", retrieved[2].Type)
}
