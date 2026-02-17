package main

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

//RPCConfig holds all configuration for RPC server

type RPCConfig struct {
	// RabbitMQ Connection
	RabbitMQ struct {
		Username       string
		Password       string
		Host           string
		Port           int
		VHost          string
		Heartbeat      time.Duration
		Timeout        time.Duration
		UseTLS         bool
		TLSConfig      *tls.Config
		MaxReconnect   int
		ReconnectDelay time.Duration
	}

	// Queue Configuration
	Queue struct {
		Name       string
		Durable    bool
		AutoDelete bool
		Exclusive  bool
		NoWait     bool
		Arguments  amqp091.Table
	}

	// Consumer Configuration
	Consumer struct {
		Tag       string
		AutoAck   bool
		Exclusive bool
		NoLocal   bool
		NoWait    bool
		Args      amqp091.Table
	}

	// QoS Configuration
	QoS struct {
		PrefetchCount int
		PrefetchSize  int
		Global        bool
	}

	// RPC Specific Configuration
	RPC struct {
		MaxWorkers     int
		ProcessTimeout time.Duration
		MaxRetries     int
		RetryDelay     time.Duration
		LogLevel       string
		EnableMetrics  bool
		MetricsPort    int
	}
}

// DefaultRPCConfig returns a production-ready default configuration
func DefaultRPCConfig() *RPCConfig {
	config := &RPCConfig{}

	// RabbitMQ Connection Defaults
	config.RabbitMQ.Username = "guest"
	config.RabbitMQ.Password = "guest"
	config.RabbitMQ.Host = "localhost"
	config.RabbitMQ.Port = 5672
	config.RabbitMQ.VHost = "/"
	config.RabbitMQ.Heartbeat = 10 * time.Second
	config.RabbitMQ.Timeout = 30 * time.Second
	config.RabbitMQ.UseTLS = false
	config.RabbitMQ.MaxReconnect = 5
	config.RabbitMQ.ReconnectDelay = 2 * time.Second

	// Queue Configuration Defaults
	config.Queue.Name = "rpc_queue"
	config.Queue.Durable = false
	config.Queue.AutoDelete = false
	config.Queue.Exclusive = false
	config.Queue.NoWait = false
	config.Queue.Arguments = amqp091.Table{}

	// Consumer Configuration Defaults
	config.Consumer.Tag = "rpc_server"
	config.Consumer.AutoAck = false
	config.Consumer.Exclusive = false
	config.Consumer.NoLocal = false
	config.Consumer.NoWait = false
	config.Consumer.Args = nil

	// QoS Defaults
	config.QoS.PrefetchCount = 1
	config.QoS.PrefetchSize = 0
	config.QoS.Global = false

	// QoS Defaults
	config.QoS.PrefetchCount = 1
	config.QoS.PrefetchSize = 0
	config.QoS.Global = false

	return config

}

// DevelopmentRPCConfig returns configuration for development
func DevelopmentRPCConfig() *RPCConfig {
	config := DefaultRPCConfig()
	config.QoS.PrefetchCount = 5
	config.RPC.LogLevel = "debug"
	config.RPC.MaxWorkers = 5
	config.RPC.ProcessTimeout = 30 * time.Second
	return config
}

// ProductionRPCConfig returns configuration for production
func ProductionRPCConfig() *RPCConfig {
	config := DefaultRPCConfig()
	config.Queue.Durable = true
	config.QoS.PrefetchCount = 10
	config.RPC.MaxWorkers = 100
	config.RPC.ProcessTimeout = 10 * time.Second
	config.RPC.MaxRetries = 5
	config.RPC.EnableMetrics = true
	config.RPC.LogLevel = "warn"
	config.RabbitMQ.MaxReconnect = 10
	config.RabbitMQ.ReconnectDelay = 5 * time.Second
	return config
}

// HighPerformanceRPCConfig for maximum throughput
func HighPerformanceRPCConfig() *RPCConfig {
	config := ProductionRPCConfig()
	config.QoS.PrefetchCount = 100
	config.RPC.MaxWorkers = 500
	config.Queue.Durable = false
	config.Queue.AutoDelete = true
	config.RPC.ProcessTimeout = 2 * time.Second
	return config
}

// LowLatencyRPCConfig for real-time systems
func LowLatencyRPCConfig() *RPCConfig {
	config := DefaultRPCConfig()
	config.QoS.PrefetchCount = 1
	config.RPC.ProcessTimeout = 500 * time.Millisecond
	config.RabbitMQ.Heartbeat = 5 * time.Second
	config.RabbitMQ.Timeout = 5 * time.Second
	return config
}

// GetConnectionURL returns the RabbitMQ connection URL
func (c *RPCConfig) GetConnectionURL() string {
	scheme := "amqp"
	if c.RabbitMQ.UseTLS {
		scheme = "amqps"
	}

	return fmt.Sprintf("%s://%s:%s@%s:%d%s",
		scheme,
		c.RabbitMQ.Username,
		c.RabbitMQ.Password,
		c.RabbitMQ.Host,
		c.RabbitMQ.Port,
		c.RabbitMQ.VHost,
	)
}

// Validate checks if configuration is valid
func (c *RPCConfig) Validate() error {
	if c.RabbitMQ.Username == "" {
		return fmt.Errorf("RabbitMQ username cannot be empty")
	}
	if c.RabbitMQ.Host == "" {
		return fmt.Errorf("RabbitMQ host cannot be empty")
	}
	if c.RabbitMQ.Port == 0 {
		return fmt.Errorf("RabbitMQ port cannot be 0")
	}
	if c.Queue.Name == "" {
		return fmt.Errorf("queue name cannot be empty")
	}
	if c.RPC.MaxWorkers <= 0 {
		return fmt.Errorf("MaxWorkers must be positive")
	}
	return nil
}
