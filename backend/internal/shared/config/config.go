package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application configuration.
type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Kafka       KafkaConfig
	WebSocket   WebSocketConfig
	Log         LogConfig
	MongoDB     MongoDBConfig
	LLMAnalyst  LLMConfig
	LLMVerifier LLMConfig
	Admin       AdminConfig
}

// AdminConfig holds administration and Agent Harness parameters.
type AdminConfig struct {
	Password             string
	DeepWikiURL          string
	RequireHumanApproval bool
}

// WebSocketConfig holds real-time WebSocket settings.
type WebSocketConfig struct {
	Enabled         bool
	MaxConnections  int
	PingIntervalSec int
	ReadBufferSize  int
	WriteBufferSize int
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host              string
	Port              int
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	CORSAllowedOrigins []string
	MaxBodyBytes      int64
}

// DatabaseConfig holds PostgreSQL connection settings.
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
	MaxConns int32
	MinConns int32
}

// DSN returns the PostgreSQL connection string with escaped credentials.
func (d DatabaseConfig) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:   "/" + d.DBName,
	}
	u.RawQuery = url.Values{"sslmode": {d.SSLMode}}.Encode()
	return u.String()
}

// RedisConfig holds Redis connection settings.
type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// Addr returns the Redis address string.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// JWTConfig holds JWT signing settings.
type JWTConfig struct {
	Secret          string
	ExpirationHours int
	Issuer          string
}

// KafkaConfig holds Kafka connection and consumer settings.
type KafkaConfig struct {
	Brokers           []string
	GroupID           string
	Enabled           bool
	NumPartitions     int
	ReplicationFactor int
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// MongoDBConfig holds MongoDB connection settings.
type MongoDBConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// LLMConfig holds configuration for a single LLM provider.
// All values are read from environment variables — nothing is hardcoded.
type LLMConfig struct {
	Provider   string        // gemma, ollama, openai, anthropic
	Endpoint   string        // e.g. http://localhost:11434, https://api.openai.com/v1
	Model      string        // e.g. gemma2:9b, gpt-4, claude-3-sonnet
	APIKey     string        // API key (optional for local models)
	Timeout    time.Duration // per-request timeout
	MaxRetries int           // retry count on transient failures
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Host:               envOrDefault("SERVER_HOST", "0.0.0.0"),
			Port:               envOrDefaultInt("SERVER_PORT", 8080),
			ReadTimeout:        time.Duration(envOrDefaultInt("SERVER_READ_TIMEOUT_SECONDS", 15)) * time.Second,
			WriteTimeout:       time.Duration(envOrDefaultInt("SERVER_WRITE_TIMEOUT_SECONDS", 15)) * time.Second,
			IdleTimeout:        time.Duration(envOrDefaultInt("SERVER_IDLE_TIMEOUT_SECONDS", 60)) * time.Second,
			CORSAllowedOrigins: envOrDefaultSlice("CORS_ALLOWED_ORIGINS", nil),
			MaxBodyBytes:       envOrDefaultInt64("MAX_BODY_BYTES", 1<<20),
		},
		Database: DatabaseConfig{
			Host:     envOrDefault("DB_HOST", "localhost"),
			Port:     envOrDefaultInt("DB_PORT", 5432),
			User:     envOrDefault("DB_USER", "masterfabric"),
			Password: envOrDefault("DB_PASSWORD", "masterfabric"),
			DBName:   envOrDefault("DB_NAME", "masterfabric"),
			SSLMode:  envOrDefault("DB_SSLMODE", "disable"),
			MaxConns: envOrDefaultInt32("DB_MAX_CONNS", 25),
			MinConns: envOrDefaultInt32("DB_MIN_CONNS", 5),
		},
		Redis: RedisConfig{
			Host:     envOrDefault("REDIS_HOST", "localhost"),
			Port:     envOrDefaultInt("REDIS_PORT", 6379),
			Password: envOrDefault("REDIS_PASSWORD", ""),
			DB:       envOrDefaultInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			Secret:          envOrDefault("JWT_SECRET", "change-me-in-production"),
			ExpirationHours: envOrDefaultInt("JWT_EXPIRATION_HOURS", 24),
			Issuer:          envOrDefault("JWT_ISSUER", "masterfabric"),
		},
		Kafka: KafkaConfig{
			Brokers:           envOrDefaultSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			GroupID:           envOrDefault("KAFKA_GROUP_ID", "masterfabric-go"),
			Enabled:           envOrDefault("KAFKA_ENABLED", "false") == "true",
			NumPartitions:     envOrDefaultInt("KAFKA_NUM_PARTITIONS", 3),
			ReplicationFactor: envOrDefaultInt("KAFKA_REPLICATION_FACTOR", 1),
		},
		WebSocket: WebSocketConfig{
			Enabled:         envOrDefault("WS_ENABLED", "true") == "true",
			MaxConnections:  envOrDefaultInt("WS_MAX_CONNECTIONS", 1000),
			PingIntervalSec: envOrDefaultInt("WS_PING_INTERVAL_SECONDS", 30),
			ReadBufferSize:  envOrDefaultInt("WS_READ_BUFFER_SIZE", 1024),
			WriteBufferSize: envOrDefaultInt("WS_WRITE_BUFFER_SIZE", 1024),
		},
		Log: LogConfig{
			Level:  envOrDefault("LOG_LEVEL", "info"),
			Format: envOrDefault("LOG_FORMAT", "json"),
		},
		MongoDB: MongoDBConfig{
			URI:      envOrDefault("MONGODB_URI", "mongodb://localhost:27017"),
			Database: envOrDefault("MONGODB_DATABASE", "claimpilot"),
			Timeout:  time.Duration(envOrDefaultInt("MONGODB_TIMEOUT_SECONDS", 10)) * time.Second,
		},
		LLMAnalyst: LLMConfig{
			Provider:   envOrDefault("LLM_ANALYST_PROVIDER", "ollama"),
			Endpoint:   envOrDefault("LLM_ANALYST_ENDPOINT", "http://localhost:11434"),
			Model:      envOrDefault("LLM_ANALYST_MODEL", "gemma2:9b"),
			APIKey:     envOrDefault("LLM_ANALYST_API_KEY", ""),
			Timeout:    time.Duration(envOrDefaultInt("LLM_ANALYST_TIMEOUT_SECONDS", 120)) * time.Second,
			MaxRetries: envOrDefaultInt("LLM_ANALYST_MAX_RETRIES", 3),
		},
		LLMVerifier: LLMConfig{
			Provider:   envOrDefault("LLM_VERIFIER_PROVIDER", "ollama"),
			Endpoint:   envOrDefault("LLM_VERIFIER_ENDPOINT", "http://localhost:11434"),
			Model:      envOrDefault("LLM_VERIFIER_MODEL", "gemma2:9b"),
			APIKey:     envOrDefault("LLM_VERIFIER_API_KEY", ""),
			Timeout:    time.Duration(envOrDefaultInt("LLM_VERIFIER_TIMEOUT_SECONDS", 60)) * time.Second,
			MaxRetries: envOrDefaultInt("LLM_VERIFIER_MAX_RETRIES", 3),
		},
		Admin: AdminConfig{
			Password:             envOrDefault("ADMIN_PASSWORD", "admin123"),
			DeepWikiURL:          envOrDefault("DEEPWIKI_MCP_URL", "http://localhost:8899/mcp/deepwiki/sse"),
			RequireHumanApproval: envOrDefault("HARNESS_REQUIRE_HUMAN_APPROVAL", "true") == "true",
		},
	}
}

func envOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func envOrDefaultInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

func envOrDefaultInt32(key string, defaultVal int32) int32 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			return int32(n)
		}
	}
	return defaultVal
}

func envOrDefaultInt64(key string, defaultVal int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return defaultVal
}

func envOrDefaultSlice(key string, defaultVal []string) []string {
	if val := os.Getenv(key); val != "" {
		parts := strings.Split(val, ",")
		var result []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultVal
}
