package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds application configuration
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	JWT       JWTConfig
	APIKeys   APIKeysConfig
	Redis     RedisConfig
	Jobs      JobsConfig
	UploadDir string // каталог для загрузки картинок товаров; пусто — загрузка отключена
}

// ServerConfig holds server settings
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database connection settings
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// JWTConfig holds JWT settings
type JWTConfig struct {
	Secret          string
	AccessTokenExp  int // minutes
	RefreshTokenExp int // hours
}

// APIKeysConfig holds external API keys (placeholders for integrations)
type APIKeysConfig struct {
	KaspiAPIKey    string
	HalykAPIKey    string
	WhatsAppAPIKey string
}

// RedisConfig кэш каталога (пустой addr — без Redis).
type RedisConfig struct {
	Addr string
}

// JobsConfig фоновые задачи.
type JobsConfig struct {
	// PendingOrderTimeoutMin через сколько минут отменять pending и снимать резерв
	PendingOrderTimeoutMin int
	// StaleSweepIntervalSec как часто запускать проверку (сек)
	StaleSweepIntervalSec int
}

// DSN returns PostgreSQL connection string
func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// Load loads configuration from environment variables
// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if exists (ignore error if not found)
	_ = godotenv.Load()

	accessExp, _ := strconv.Atoi(getEnv("JWT_ACCESS_EXP", "60"))
	refreshExp, _ := strconv.Atoi(getEnv("JWT_REFRESH_EXP", "168"))
	pendingTimeout, _ := strconv.Atoi(getEnv("PENDING_ORDER_TIMEOUT_MIN", "120"))
	staleSweep, _ := strconv.Atoi(getEnv("STALE_ORDER_SWEEP_INTERVAL_SEC", "300"))

	// Если DATABASE_URL есть (Railway), используй его напрямую
	databaseURL := os.Getenv("DATABASE_URL")

	var dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode string

	if databaseURL != "" {
		// Парсим DATABASE_URL (формат: postgresql://user:password@host:port/dbname?sslmode=...)
		// Для простоты используем дефолты из отдельных переменных если есть
		dbHost = getEnv("PGHOST", "localhost")
		dbPort = getEnv("PGPORT", "5432")
		dbUser = getEnv("PGUSER", "postgres")
		dbPassword = getEnv("PGPASSWORD", "password")
		dbName = getEnv("PGDATABASE", "veggies_shop")
		dbSSLMode = getEnv("PGSSLMODE", "require") // Railway требует SSL
	} else {
		// Локальная разработка
		dbHost = getEnv("DB_HOST", "localhost")
		dbPort = getEnv("DB_PORT", "5432")
		dbUser = getEnv("DB_USER", "postgres")
		dbPassword = getEnv("DB_PASSWORD", "password")
		dbName = getEnv("DB_NAME", "veggies_shop")
		dbSSLMode = getEnv("DB_SSLMODE", "disable")
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Name:     dbName,
			SSLMode:  dbSSLMode,
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			AccessTokenExp:  accessExp,
			RefreshTokenExp: refreshExp,
		},
		APIKeys: APIKeysConfig{
			KaspiAPIKey:    getEnv("KASPI_API_KEY", ""),
			HalykAPIKey:    getEnv("HALYK_API_KEY", ""),
			WhatsAppAPIKey: getEnv("WHATSAPP_API_KEY", ""),
		},
		Redis: RedisConfig{
			Addr: getEnv("REDIS_ADDR", ""),
		},
		Jobs: JobsConfig{
			PendingOrderTimeoutMin: pendingTimeout,
			StaleSweepIntervalSec:  staleSweep,
		},
		UploadDir: getEnv("UPLOAD_DIR", "data/uploads"),
	}

	return cfg, nil
}

// InitLogger initializes structured logging
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	slog.SetDefault(slog.New(handler))
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
