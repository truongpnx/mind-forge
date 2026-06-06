package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/redis/go-redis/v9"
)

// PostgresConfig holds connection parameters for PostgreSQL.
// ConnectionString takes priority: when non-empty it is returned by DSN() as-is.
type PostgresConfig struct {
	ConnectionString string `toml:"connection_string"`
	PostgresHost     string `toml:"postgres_host"`
	Username         string `toml:"username"`
	Password         string `toml:"password"`
	Port             int    `toml:"port"`
	Database         string `toml:"database"`
}

// DSN returns the connection string. If ConnectionString is set it is returned
// directly; otherwise a URL is built from the individual fields.
func (c PostgresConfig) DSN() string {
	if c.ConnectionString != "" {
		return c.ConnectionString
	}
	port := c.Port
	if port == 0 {
		port = 5432
	}
	db := c.Database
	if db == "" {
		db = "db"
	}
	host := c.PostgresHost
	if host == "" {
		host = "localhost"
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.Username, c.Password, host, port, db,
	)
}

// RedisConfig holds connection parameters for Redis.
// Socket takes priority over host/port when non-empty (Unix domain socket path).
type RedisConfig struct {
	Socket   string `toml:"socket"`
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// Options returns a *redis.Options suitable for redis.NewClient.
func (c RedisConfig) Options() *redis.Options {
	opts := &redis.Options{
		Username: c.Username,
		Password: c.Password,
	}
	if c.Socket != "" {
		opts.Network = "unix"
		opts.Addr = c.Socket
		return opts
	}
	host := c.Host
	if host == "" {
		host = "localhost"
	}
	port := c.Port
	if port == 0 {
		port = 6379
	}
	opts.Addr = fmt.Sprintf("%s:%d", host, port)
	return opts
}

// Config is the root configuration structure.
type Config struct {
	PostgreSQL PostgresConfig `toml:"postgresql"`
	Redis      RedisConfig    `toml:"redis"`
}

// Load reads a TOML config file from path.
// The path can be overridden by the CONFIG_FILE environment variable.
func Load(path string) (*Config, error) {
	if envPath := os.Getenv("CONFIG_FILE"); envPath != "" {
		path = envPath
	}
	var cfg Config
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("config: decode %q: %w", path, err)
	}
	return &cfg, nil
}
