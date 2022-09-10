package config

import (
	"net"
	"time"
)

type AgentConfig struct {
	ServerAddress  string
	Key            string
	PollInterval   time.Duration
	ReportInterval time.Duration
	PublicKeyPath  string
}

type AgentConfigJSON struct {
	ServerAddress  string `json:"address,omitempty"`
	PollInterval   string `json:"poll_interval,omitempty"`
	ReportInterval string `json:"report_interval,omitempty"`
	PublicKeyPath  string `json:"crypto_key,omitempty"`
	Key            string `json:"hash_key,omitempty"`
}

type ServerConfig struct {
	ServerAddress  string
	StoreFile      string
	Key            string
	DSN            string
	StoreInterval  time.Duration
	UseDB          bool
	Restore        bool
	PrivateKeyPath string
	TrustedSubnet  net.IPNet
}

type ServerConfigJSON struct {
	ServerAddress  string `json:"address,omitempty"`
	StoreFile      string `json:"store_file,omitempty"`
	Key            string `json:"hash_key,omitempty"`
	DSN            string `json:"database_dsn,omitempty"`
	StoreInterval  string `json:"store_interval,omitempty"`
	Restore        *bool  `json:"restore,omitempty"`
	PrivateKeyPath string `json:"crypto_key,omitempty"`
	TrustedSubnet  string `json:"trusted_subnet,omitempty"`
}
