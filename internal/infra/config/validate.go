package config

import (
	"fmt"
	"net"
	"time"
)

func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("validate runtime config: config is nil")
	}

	if err := validateHTTP(&cfg.HTTP); err != nil {
		return fmt.Errorf("failed validate http config: %w", err)
	}

	return nil
}

func validateHTTP(httpCfg *HTTP) error {
	if httpCfg == nil {
		return fmt.Errorf("http: configuration is nil")
	}

	if _, _, err := net.SplitHostPort(httpCfg.Addr); err != nil {
		return fmt.Errorf("invalid addr: %w", err)
	}

	if httpCfg.Timeouts.ReadTimeout < time.Millisecond {
		return fmt.Errorf("read_timeout must be at least 1ms")
	}

	if httpCfg.Timeouts.ReadHeaderTimeout > httpCfg.Timeouts.ReadTimeout {
		return fmt.Errorf("read_header_timeout cannot exceed read_timeout")
	}

	if httpCfg.Timeouts.WriteTimeout < 0 {
		return fmt.Errorf("write_timeout cannot be negative")
	}

	if httpCfg.MaxHeaderBytes < 0 {
		return fmt.Errorf("max_header_bytes cannot be negative")
	}

	return nil
}
