package persistence

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestIPDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&IPAddress{})
	require.NoError(t, err)

	return db
}

func TestIPRepo_AllocateIP(t *testing.T) {
	db := setupTestIPDB(t)
	repo := NewIPRepo(db)
	ctx := context.Background()

	// Create some test IP addresses
	ips := []*IPAddress{
		{Address: "192.168.1.1", Allocated: false},
		{Address: "192.168.1.2", Allocated: false},
		{Address: "192.168.1.3", Allocated: true}, // Already allocated
	}

	for _, ip := range ips {
		err := db.Create(ip).Error
		require.NoError(t, err)
	}

	// Test allocation
	allocated, err := repo.AllocateIP(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, allocated)
	assert.Equal(t, "192.168.1.1", allocated.Address)
	assert.True(t, allocated.Allocated)

	// Test second allocation
	allocated2, err := repo.AllocateIP(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, allocated2)
	assert.Equal(t, "192.168.1.2", allocated2.Address)
	assert.True(t, allocated2.Allocated)

	// Test no more available IPs
	allocated3, err := repo.AllocateIP(ctx)
	assert.NoError(t, err)
	assert.Nil(t, allocated3) // No more available IPs
}

func TestIPRepo_ReleaseIP(t *testing.T) {
	db := setupTestIPDB(t)
	repo := NewIPRepo(db)
	ctx := context.Background()

	// Create an allocated IP
	ip := &IPAddress{
		Address:   "192.168.1.1",
		Allocated: true,
		ServerID:  stringPtr("server-123"),
	}
	err := db.Create(ip).Error
	require.NoError(t, err)

	// Release the IP
	err = repo.ReleaseIP(ctx, ip.ID)
	assert.NoError(t, err)

	// Verify it was released
	var found IPAddress
	err = db.First(&found, ip.ID).Error
	assert.NoError(t, err)
	assert.False(t, found.Allocated)
	assert.Nil(t, found.ServerID)
}

func TestIPRepo_AssignIPToServer(t *testing.T) {
	db := setupTestIPDB(t)
	repo := NewIPRepo(db)
	ctx := context.Background()

	// Create an unallocated IP
	ip := &IPAddress{
		Address:   "192.168.1.1",
		Allocated: false,
	}
	err := db.Create(ip).Error
	require.NoError(t, err)

	// Assign to server
	err = repo.AssignIPToServer(ctx, ip.ID, "server-123")
	assert.NoError(t, err)

	// Verify assignment
	var found IPAddress
	err = db.First(&found, ip.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "server-123", *found.ServerID)
}

func TestIPRepo_ConcurrentAllocation(t *testing.T) {
	db := setupTestIPDB(t)
	repo := NewIPRepo(db)
	ctx := context.Background()

	// Create multiple unallocated IPs
	for i := 1; i <= 10; i++ {
		ip := &IPAddress{
			Address:   "192.168.1." + string(rune(i)),
			Allocated: false,
		}
		err := db.Create(ip).Error
		require.NoError(t, err)
	}

	// Test concurrent allocation
	results := make(chan *IPAddress, 10)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			ip, err := repo.AllocateIP(ctx)
			if err != nil {
				errors <- err
				return
			}
			results <- ip
		}()
	}

	// Collect results
	allocated := 0
	for i := 0; i < 10; i++ {
		select {
		case ip := <-results:
			assert.NotNil(t, ip)
			assert.True(t, ip.Allocated)
			allocated++
		case err := <-errors:
			assert.NoError(t, err)
		}
	}

	assert.Equal(t, 10, allocated)

	// Verify no more IPs are available
	remaining, err := repo.AllocateIP(ctx)
	assert.NoError(t, err)
	assert.Nil(t, remaining)
}

func TestIPRepo_AllocateIP_NoAvailable(t *testing.T) {
	db := setupTestIPDB(t)
	repo := NewIPRepo(db)
	ctx := context.Background()

	// Create only allocated IPs
	ip := &IPAddress{
		Address:   "192.168.1.1",
		Allocated: true,
	}
	err := db.Create(ip).Error
	require.NoError(t, err)

	// Try to allocate
	allocated, err := repo.AllocateIP(ctx)
	assert.NoError(t, err)
	assert.Nil(t, allocated) // No available IPs
}

func TestIPRepo_ReleaseIP_NonExistent(t *testing.T) {
	db := setupTestIPDB(t)
	repo := NewIPRepo(db)
	ctx := context.Background()

	// Try to release non-existent IP
	err := repo.ReleaseIP(ctx, 999)
	assert.NoError(t, err) // GORM doesn't error on non-existent updates
}

func TestIPRepo_AssignIPToServer_NonExistent(t *testing.T) {
	db := setupTestIPDB(t)
	repo := NewIPRepo(db)
	ctx := context.Background()

	// Try to assign non-existent IP
	err := repo.AssignIPToServer(ctx, 999, "server-123")
	assert.Error(t, err) // GORM will error on non-existent record
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
