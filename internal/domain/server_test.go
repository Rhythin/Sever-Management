package domain

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServer_NewEventRingBuffer(t *testing.T) {
	buf := NewEventRingBuffer(10)
	assert.NotNil(t, buf)
	assert.Equal(t, 10, buf.size)
	assert.Equal(t, 0, buf.count)
}

func TestServer_Transition_ValidTransitions(t *testing.T) {
	server := &Server{
		ID:     "test-server",
		Type:   TypeT2Micro,
		Region: "us-west-1",
		State:  ServerStopped,
		Log:    NewEventRingBuffer(10),
	}
	ctx := context.Background()

	// Test valid transitions
	tests := []struct {
		from     ServerState
		action   ServerAction
		expected ServerState
	}{
		{ServerStopped, ActionStart, ServerRunning},
		{ServerRunning, ActionStop, ServerStopped},
		{ServerRunning, ActionReboot, ServerRunning},
		{ServerStopped, ActionTerminate, ServerTerminated},
		{ServerRunning, ActionTerminate, ServerTerminated},
	}

	for _, tt := range tests {
		server.State = tt.from
		err := server.Transition(ctx, tt.action)
		assert.NoError(t, err)
		assert.Equal(t, tt.expected, server.State)
	}
}

func TestServer_Transition_InvalidTransitions(t *testing.T) {
	server := &Server{
		ID:     "test-server",
		Type:   TypeT2Micro,
		Region: "us-west-1",
		State:  ServerTerminated,
		Log:    NewEventRingBuffer(10),
	}
	ctx := context.Background()

	// Test invalid transitions
	tests := []struct {
		from   ServerState
		action ServerAction
	}{
		{ServerTerminated, ActionStart},
		{ServerTerminated, ActionStop},
		{ServerTerminated, ActionReboot},
		{ServerTerminated, ActionTerminate},
	}

	for _, tt := range tests {
		server.State = tt.from
		err := server.Transition(ctx, tt.action)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid state transition")
	}
}

func TestServer_Transition_EventLogging(t *testing.T) {
	server := &Server{
		ID:     "test-server",
		Type:   TypeT2Micro,
		Region: "us-west-1",
		State:  ServerStopped,
		Log:    NewEventRingBuffer(10),
	}
	ctx := context.Background()

	// Test that events are logged
	err := server.Transition(ctx, ActionStart)
	require.NoError(t, err)

	events := server.Log.List()
	assert.Len(t, events, 1)
	assert.Equal(t, EventStarted, events[0].Type)
	assert.Contains(t, events[0].Message, "started")
}

func TestServer_Transition_TimestampUpdates(t *testing.T) {
	server := &Server{
		ID:     "test-server",
		Type:   TypeT2Micro,
		Region: "us-west-1",
		State:  ServerStopped,
		Log:    NewEventRingBuffer(10),
	}
	ctx := context.Background()

	// Test start action updates StartedAt
	err := server.Transition(ctx, ActionStart)
	require.NoError(t, err)
	assert.NotNil(t, server.StartedAt)
	assert.Nil(t, server.StoppedAt)

	// Test stop action updates StoppedAt
	err = server.Transition(ctx, ActionStop)
	require.NoError(t, err)
	assert.NotNil(t, server.StoppedAt)

	// Test terminate action updates TerminatedAt
	err = server.Transition(ctx, ActionTerminate)
	require.NoError(t, err)
	assert.NotNil(t, server.TerminatedAt)
}

func TestServer_EventRingBuffer(t *testing.T) {
	server := &Server{
		ID:     "test-server",
		Type:   TypeT2Micro,
		Region: "us-west-1",
		State:  ServerStopped,
		Log:    NewEventRingBuffer(3), // Small buffer for testing
	}
	ctx := context.Background()

	// Add more than 3 events to test ring buffer behavior
	for i := 0; i < 5; i++ {
		server.Transition(ctx, ActionStart)
		server.Transition(ctx, ActionStop)
	}

	// Should only keep last 3 events
	events := server.Log.List()
	assert.Len(t, events, 3)

	// Latest event should be the last one
	latest := events[len(events)-1]
	assert.Equal(t, EventStopped, latest.Type)
}

func TestServer_ConcurrentAccess(t *testing.T) {
	server := &Server{
		ID:     "test-server",
		Type:   TypeT2Micro,
		Region: "us-west-1",
		State:  ServerStopped,
		Log:    NewEventRingBuffer(10),
	}
	ctx := context.Background()

	// Test concurrent transitions with alternating actions
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			// Alternate between start and stop actions
			var action ServerAction
			if id%2 == 0 {
				action = ActionStart
			} else {
				action = ActionStop
			}

			err := server.Transition(ctx, action)
			// Some transitions might fail due to invalid state, which is expected
			// We're testing that the server doesn't crash under concurrent access
			_ = err // Ignore errors for this test
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// State should be consistent (either running or stopped, not corrupted)
	assert.Contains(t, []ServerState{ServerRunning, ServerStopped}, server.State)
}

func TestEventRingBuffer_AddAndList(t *testing.T) {
	buf := NewEventRingBuffer(3)
	now := time.Now()

	// Add events
	event1 := EventLogEntry{Timestamp: now, Type: EventStarted, Message: "started"}
	event2 := EventLogEntry{Timestamp: now.Add(time.Second), Type: EventStopped, Message: "stopped"}
	event3 := EventLogEntry{Timestamp: now.Add(2 * time.Second), Type: EventStarted, Message: "started again"}

	buf.Add(event1)
	buf.Add(event2)
	buf.Add(event3)

	events := buf.List()
	assert.Len(t, events, 3)
	assert.Equal(t, event1, events[0])
	assert.Equal(t, event2, events[1])
	assert.Equal(t, event3, events[2])
}

func TestEventRingBuffer_Overflow(t *testing.T) {
	buf := NewEventRingBuffer(2)
	now := time.Now()

	// Add 3 events to a buffer of size 2
	event1 := EventLogEntry{Timestamp: now, Type: EventStarted, Message: "first"}
	event2 := EventLogEntry{Timestamp: now.Add(time.Second), Type: EventStopped, Message: "second"}
	event3 := EventLogEntry{Timestamp: now.Add(2 * time.Second), Type: EventStarted, Message: "third"}

	buf.Add(event1)
	buf.Add(event2)
	buf.Add(event3)

	events := buf.List()
	assert.Len(t, events, 2)
	// Should contain the last 2 events
	assert.Equal(t, event2, events[0])
	assert.Equal(t, event3, events[1])
}

func TestServerState_String(t *testing.T) {
	assert.Equal(t, "stopped", string(ServerStopped))
	assert.Equal(t, "running", string(ServerRunning))
	assert.Equal(t, "terminated", string(ServerTerminated))
}

func TestServerAction_String(t *testing.T) {
	assert.Equal(t, "start", string(ActionStart))
	assert.Equal(t, "stop", string(ActionStop))
	assert.Equal(t, "reboot", string(ActionReboot))
	assert.Equal(t, "terminate", string(ActionTerminate))
}

func TestIsValidAction(t *testing.T) {
	type args struct {
		action ServerAction
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "valid action",
			args: args{
				action: ActionStart,
			},
			want: true,
		},
		{
			name: "invalid action",
			args: args{
				action: ServerAction("invalid-action"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidAction(tt.args.action); got != tt.want {
				t.Errorf("IsValidAction() = %v, want %v", got, tt.want)
			}
		})
	}
}
