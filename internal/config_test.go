package internal

import (
    "os"
    "testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
    os.Clearenv()
    cfg, err := LoadConfig()
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if cfg.HTTPPort != 8080 {
        t.Errorf("expected default HTTPPort 8080, got %d", cfg.HTTPPort)
    }
    if cfg.Env != "development" {
        t.Errorf("expected default Env 'development', got %s", cfg.Env)
    }
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
    os.Clearenv()
    os.Setenv("HTTP_PORT", "1234")
    os.Setenv("ENV", "production")
    cfg, err := LoadConfig()
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if cfg.HTTPPort != 1234 {
        t.Errorf("expected HTTPPort 1234, got %d", cfg.HTTPPort)
    }
    if cfg.Env != "production" {
        t.Errorf("expected Env 'production', got %s", cfg.Env)
    }
}
