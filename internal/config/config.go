package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	WebSocket WebSocketConfig
}

type ServerConfig struct {
	Port            string        `env:"SERVER_PORT" envDefault:"8080"`
	ReadTimeout     time.Duration `env:"SERVER_READ_TIMEOUT" envDefault:"10s"`
	WriteTimeout    time.Duration `env:"SERVER_WRITE_TIMEOUT" envDefault:"10s"`
	ShutdownTimeout time.Duration `env:"SERVER_SHUTDOWN_TIMEOUT" envDefault:"15s"`
}

type DatabaseConfig struct {
	DSN             string        `env:"DATABASE_DSN,required"`
	MaxConns        int32         `env:"DATABASE_MAX_CONNS" envDefault:"25"`
	MinConns        int32         `env:"DATABASE_MIN_CONNS" envDefault:"5"`
	MaxConnLifetime time.Duration `env:"DATABASE_MAX_CONN_LIFETIME" envDefault:"1h"`
	MaxConnIdleTime time.Duration `env:"DATABASE_MAX_CONN_IDLE_TIME" envDefault:"30m"`
}

type WebSocketConfig struct {
	PingInterval     time.Duration `env:"WS_PING_INTERVAL" envDefault:"30s"`
	PongWait         time.Duration `env:"WS_PONG_WAIT" envDefault:"60s"`
	WriteWait        time.Duration `env:"WS_WRITE_WAIT" envDefault:"10s"`
	MaxMessageSize   int64         `env:"WS_MAX_MESSAGE_SIZE" envDefault:"8192"`
	ReadBufferSize   int           `env:"WS_READ_BUFFER_SIZE" envDefault:"4096"`
	WriteBufferSize  int           `env:"WS_WRITE_BUFFER_SIZE" envDefault:"4096"`
	BroadcastBuffer  int           `env:"WS_BROADCAST_BUFFER" envDefault:"256"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
