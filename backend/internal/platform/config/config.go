package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"
)

const minimumJWTSecretLength = 32

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
	EnvironmentTest        = "test"
)

type DatabaseConfig struct {
	URL            string
	MaxConnections int32
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

type SessionConfig struct {
	JWTSecret         string
	IsCookieSecure    bool
	TOTPEncryptionKey string
}

type Config struct {
	Environment     string
	HTTPPort        int
	LogLevel        slog.Level
	ShutdownTimeout time.Duration
	AllowedOrigins  []string
	TrustedProxies  []string
	RedisURL        string
	FrontendBaseURL string
	Database        DatabaseConfig
	Email           EmailConfig
	Session         SessionConfig
	Stripe          StripeConfig
}

type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
}

func (stripeConfig StripeConfig) IsConfigured() bool {
	return stripeConfig.SecretKey != ""
}

func loadStripeConfig() (StripeConfig, error) {
	stripeConfig := StripeConfig{
		SecretKey:     readString("STRIPE_SECRET_KEY", ""),
		WebhookSecret: readString("STRIPE_WEBHOOK_SECRET", ""),
	}
	if stripeConfig.SecretKey != "" && stripeConfig.WebhookSecret == "" {
		return StripeConfig{}, fmt.Errorf("STRIPE_WEBHOOK_SECRET is required when STRIPE_SECRET_KEY is set")
	}
	return stripeConfig, nil
}

func Load() (Config, error) {
	environment := readString("APP_ENV", EnvironmentDevelopment)
	if !isSupportedEnvironment(environment) {
		return Config{}, fmt.Errorf("APP_ENV %q is not one of development, production, test", environment)
	}

	httpPort, err := readInt("HTTP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	logLevel, err := readLogLevel("LOG_LEVEL", slog.LevelInfo)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := readDuration("SHUTDOWN_TIMEOUT", 15*time.Second)
	if err != nil {
		return Config{}, err
	}

	allowedOrigins := readList("FRONTEND_ORIGINS", []string{"http://localhost:3000"})

	trustedProxies := readList("TRUSTED_PROXIES", nil)
	if err := validateTrustedProxies(trustedProxies); err != nil {
		return Config{}, err
	}

	databaseConfig, err := loadDatabaseConfig()
	if err != nil {
		return Config{}, err
	}

	emailConfig, err := loadEmailConfig()
	if err != nil {
		return Config{}, err
	}

	sessionConfig, err := loadSessionConfig(environment)
	if err != nil {
		return Config{}, err
	}

	stripeConfig, err := loadStripeConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:     environment,
		HTTPPort:        httpPort,
		LogLevel:        logLevel,
		ShutdownTimeout: shutdownTimeout,
		AllowedOrigins:  allowedOrigins,
		TrustedProxies:  trustedProxies,
		RedisURL:        readString("REDIS_URL", "redis://localhost:6379/0"),
		FrontendBaseURL: readString("FRONTEND_BASE_URL", "http://localhost:3000"),
		Database:        databaseConfig,
		Email:           emailConfig,
		Session:         sessionConfig,
		Stripe:          stripeConfig,
	}, nil
}

func loadSessionConfig(environment string) (SessionConfig, error) {
	jwtSecret := readString("JWT_SECRET", "")
	if len(jwtSecret) < minimumJWTSecretLength {
		return SessionConfig{}, fmt.Errorf("JWT_SECRET must be at least %d characters", minimumJWTSecretLength)
	}
	isCookieSecure, err := readBool("COOKIE_SECURE", environment != EnvironmentDevelopment)
	if err != nil {
		return SessionConfig{}, err
	}
	totpEncryptionKey := readString("TOTP_ENCRYPTION_KEY", "")
	if totpEncryptionKey == "" {
		return SessionConfig{}, errors.New("TOTP_ENCRYPTION_KEY is required")
	}
	return SessionConfig{JWTSecret: jwtSecret, IsCookieSecure: isCookieSecure, TOTPEncryptionKey: totpEncryptionKey}, nil
}

func loadEmailConfig() (EmailConfig, error) {
	smtpPort, err := readInt("SMTP_PORT", 1025)
	if err != nil {
		return EmailConfig{}, err
	}
	return EmailConfig{
		SMTPHost:     readString("SMTP_HOST", "localhost"),
		SMTPPort:     smtpPort,
		SMTPUsername: readString("SMTP_USERNAME", ""),
		SMTPPassword: readString("SMTP_PASSWORD", ""),
		FromAddress:  readString("EMAIL_FROM_ADDRESS", "no-reply@nebula.test"),
		FromName:     readString("EMAIL_FROM_NAME", "Nebula Exchange"),
	}, nil
}

func loadDatabaseConfig() (DatabaseConfig, error) {
	databaseURL := readString("DATABASE_URL", "")
	if databaseURL == "" {
		return DatabaseConfig{}, errors.New("DATABASE_URL is required")
	}

	maxConnections, err := readInt("DATABASE_MAX_CONNECTIONS", 20)
	if err != nil {
		return DatabaseConfig{}, err
	}
	if maxConnections < 1 {
		return DatabaseConfig{}, errors.New("DATABASE_MAX_CONNECTIONS must be at least 1")
	}

	return DatabaseConfig{URL: databaseURL, MaxConnections: int32(maxConnections)}, nil
}

func (cfg Config) IsProduction() bool {
	return cfg.Environment == EnvironmentProduction
}

func (cfg Config) HTTPAddress() string {
	return fmt.Sprintf(":%d", cfg.HTTPPort)
}

func isSupportedEnvironment(environment string) bool {
	switch environment {
	case EnvironmentDevelopment, EnvironmentProduction, EnvironmentTest:
		return true
	default:
		return false
	}
}

func validateTrustedProxies(trustedProxies []string) error {
	for _, trustedProxy := range trustedProxies {
		if net.ParseIP(trustedProxy) != nil {
			continue
		}
		if _, _, err := net.ParseCIDR(trustedProxy); err != nil {
			return fmt.Errorf("TRUSTED_PROXIES entry %q is not an IP address or CIDR range", trustedProxy)
		}
	}
	return nil
}
