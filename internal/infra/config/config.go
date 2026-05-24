package config

import (
	"crypto/tls"
	"time"
)

type HTTP struct {
	Addr           string     `yaml:"addr"`
	TLS            tls.Config `yaml:"tls"`
	Timeouts       Timeouts   `yaml:"timeouts"`
	MaxHeaderBytes int        `yaml:"max_header_bytes"`
}

type Timeouts struct {
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	IdleTimeout       time.Duration `yaml:"idle_timeout"`
}

type KafkaConfig struct {
	Brokers         []string `yaml:"brokers"`
	Topic           string   `yaml:"topic"`
	GroupID         string   `yaml:"group_id"`
	ClientID        string   `yaml:"client_id"`
	AutoOffsetReset string   `yaml:"auto_offset_reset"`
}

type Config struct {
	HTTP        HTTP        `yaml:"http"`
	KafkaConfig KafkaConfig `yaml:"kafka"`
}
