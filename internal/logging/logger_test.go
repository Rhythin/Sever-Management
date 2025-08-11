package logging

import (
	"testing"

	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	logger := InitLogger()
	if logger == nil {
		t.Fatal("InitLogger returned nil")
	}
	// Verify global logger matches returned logger
	global := zap.L()
	if global == nil {
		t.Fatal("Global logger is nil after InitLogger")
	}
	if logger.Core() == nil || global.Core() == nil {
		t.Error("Logger core is nil")
	}
}
