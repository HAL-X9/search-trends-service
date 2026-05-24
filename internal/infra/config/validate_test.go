package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	validCfg := &Config{
		HTTP: HTTP{
			Addr: ":8080",
			Timeouts: Timeouts{
				ReadTimeout:       time.Second,
				ReadHeaderTimeout: time.Millisecond * 500,
				WriteTimeout:      time.Second,
			},
			MaxHeaderBytes: 1024,
		},
	}

	tests := []struct {
		name    string
		cfg     *Config
		wantErr string
	}{
		{
			name:    "nil config",
			cfg:     nil,
			wantErr: "config is nil",
		},
		{
			name: "invalid addr",
			cfg: func() *Config {
				cfg := *validCfg
				cfg.HTTP.Addr = "invalid"
				return &cfg
			}(),
			wantErr: "invalid addr",
		},
		{
			name: "read timeout less than 1ms",
			cfg: func() *Config {
				cfg := *validCfg
				cfg.HTTP.Timeouts.ReadTimeout = 0
				return &cfg
			}(),
			wantErr: "read_timeout must be at least 1ms",
		},
		{
			name: "read header timeout exceeds read timeout",
			cfg: func() *Config {
				cfg := *validCfg
				cfg.HTTP.Timeouts.ReadTimeout = time.Second
				cfg.HTTP.Timeouts.ReadHeaderTimeout = 2 * time.Second
				return &cfg
			}(),
			wantErr: "read_header_timeout cannot exceed read_timeout",
		},
		{
			name: "negative write timeout",
			cfg: func() *Config {
				cfg := *validCfg
				cfg.HTTP.Timeouts.WriteTimeout = -time.Second
				return &cfg
			}(),
			wantErr: "write_timeout cannot be negative",
		},
		{
			name: "negative max header bytes",
			cfg: func() *Config {
				cfg := *validCfg
				cfg.HTTP.MaxHeaderBytes = -1
				return &cfg
			}(),
			wantErr: "max_header_bytes cannot be negative",
		},
		{
			name: "valid config",
			cfg:  validCfg,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := Validate(tt.cfg)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("expected nil error, got: %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
