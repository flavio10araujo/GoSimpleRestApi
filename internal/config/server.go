package config

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

type ServerConfig struct {
	Host              string
	Port              string
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	ReadHeaderTimeout time.Duration
}

func LoadServerConfig() (*ServerConfig, error) {
	host := GetEnv("HOST", "")
	port := GetEnv("PORT", "8080")
	if err := validatePort(port); err != nil {
		return nil, fmt.Errorf("invalid PORT: %w", err)
	}

	readHeaderTimeout, err := parseDurationEnv("READ_HEADER_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid READ_HEADER_TIMEOUT: %w", err)
	}
	readTimeout, err := parseDurationEnv("READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid READ_TIMEOUT: %w", err)
	}
	writeTimeout, err := parseDurationEnv("WRITE_TIMEOUT", 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid WRITE_TIMEOUT: %w", err)
	}
	idleTimeout, err := parseDurationEnv("IDLE_TIMEOUT", 60*time.Second)
	if err != nil {
		return nil, fmt.Errorf("invalid IDLE_TIMEOUT: %w", err)
	}

	return &ServerConfig{
		Host:              host,
		Port:              port,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
	}, nil
}

func (c *ServerConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

func parseDurationEnv(key string, fallback time.Duration) (time.Duration, error) {
	value := GetEnv(key, fallback.String())
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

func validatePort(port string) error {
	value, err := strconv.Atoi(port)
	if err != nil {
		return err
	}
	if value < 1 || value > 65535 {
		return fmt.Errorf("must be between 1 and 65535")
	}

	return nil
}
