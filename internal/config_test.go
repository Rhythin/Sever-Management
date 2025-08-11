package internal

import (
	"reflect"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    *Config
		wantErr bool
	}{
		{
			name: "valid config",
			want: &Config{
				Env:              "development",
				HTTPPort:         8080,
				DBHost:           "localhost",
				DBPort:           5432,
				DBUser:           "postgres",
				DBPassword:       "password",
				DBName:           "servermgmt",
				DBSSLMode:        "disable",
				BillingRate:      0.01,
				IdleTimeout:      30 * time.Minute,
				BillingInterval:  time.Minute,
				ReaperInterval:   5 * time.Minute,
				EnableIdleReaper: true,
				IPCIDR:           "192.168.0.0/16",
				LogLevel:         "info",
				MetricsPort:      9090,
				RequestTimeout:   30 * time.Second,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
