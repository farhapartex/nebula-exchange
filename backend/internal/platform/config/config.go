package config

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"
)

const MinimumJWTSecretLength = 32

const (
	defaultChainID               = 31337
	defaultRequiredConfirmations = 1
	defaultChainPollInterval     = 3 * time.Second
)

const (
	EnvironmentDevelopment = "development"
	EnvironmentProduction  = "production"
	EnvironmentTest        = "test"
)

type DatabaseConfig struct {
	URL                   string
	MaxOpenConnections    int
	MaxIdleConnections    int
	ConnectionMaxLifetime time.Duration
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

type StorageConfig struct {
	Endpoint        string
	PublicEndpoint  string
	AccessKey       string
	SecretKey       string
	Bucket          string
	Region          string
	UseSSL          bool
	PresignLifetime time.Duration
}

type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
}

func (stripeConfig StripeConfig) IsConfigured() bool {
	return stripeConfig.SecretKey != "" && stripeConfig.WebhookSecret != ""
}

type WalletPaymentConfig struct {
	RPCURL                     string
	ChapterPaymentVaultAddress string
	PaymentSignerPrivateKey    string
	RequiredConfirmations      int
	PollInterval               time.Duration
}

func (walletPaymentConfig WalletPaymentConfig) IsConfigured() bool {
	return walletPaymentConfig.RPCURL != "" && walletPaymentConfig.ChapterPaymentVaultAddress != "" && walletPaymentConfig.PaymentSignerPrivateKey != ""
}

type SessionConfig struct {
	JWTSecret      string
	IsCookieSecure bool
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
	Storage         StorageConfig
	Stripe          StripeConfig
	ChainID         int64
	WalletPayment   WalletPaymentConfig
}

func Load() (Config, error) {
	environment, err := loadEnvironment()
	if err != nil {
		return Config{}, err
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

	trustedProxies := readList("TRUSTED_PROXIES", nil)
	if err := validateTrustedProxies(trustedProxies); err != nil {
		return Config{}, err
	}

	databaseConfig, err := LoadDatabaseConfig()
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

	storageConfig, err := LoadStorageConfig()
	if err != nil {
		return Config{}, err
	}

	chainID, err := readInt("CHAIN_ID", defaultChainID)
	if err != nil {
		return Config{}, err
	}

	walletPaymentConfig, err := loadWalletPaymentConfig()
	if err != nil {
		return Config{}, err
	}

	return Config{
		Environment:     environment,
		HTTPPort:        httpPort,
		LogLevel:        logLevel,
		ShutdownTimeout: shutdownTimeout,
		AllowedOrigins:  readList("FRONTEND_ORIGINS", []string{"http://localhost:3000"}),
		TrustedProxies:  trustedProxies,
		RedisURL:        readString("REDIS_URL", "redis://localhost:6379/0"),
		FrontendBaseURL: readString("FRONTEND_BASE_URL", "http://localhost:3000"),
		Database:        databaseConfig,
		Email:           emailConfig,
		Session:         sessionConfig,
		Storage:         storageConfig,
		ChainID:         int64(chainID),
		WalletPayment:   walletPaymentConfig,
		Stripe: StripeConfig{
			SecretKey:     readString("STRIPE_SECRET_KEY", ""),
			WebhookSecret: readString("STRIPE_WEBHOOK_SECRET", ""),
		},
	}, nil
}

func LoadStorageConfig() (StorageConfig, error) {
	accessKey := readString("STORAGE_ACCESS_KEY", "")
	secretKey := readString("STORAGE_SECRET_KEY", "")
	if accessKey == "" || secretKey == "" {
		return StorageConfig{}, errors.New("STORAGE_ACCESS_KEY and STORAGE_SECRET_KEY are required")
	}
	useSSL, err := readBool("STORAGE_USE_SSL", false)
	if err != nil {
		return StorageConfig{}, err
	}
	presignLifetime, err := readDuration("STORAGE_PRESIGN_LIFETIME", time.Hour)
	if err != nil {
		return StorageConfig{}, err
	}
	if presignLifetime < time.Minute || presignLifetime > 7*24*time.Hour {
		return StorageConfig{}, errors.New("STORAGE_PRESIGN_LIFETIME must be between 1m and 168h")
	}
	endpoint := readString("STORAGE_ENDPOINT", "localhost:9000")
	return StorageConfig{
		Endpoint:        endpoint,
		PublicEndpoint:  readString("STORAGE_PUBLIC_ENDPOINT", endpoint),
		AccessKey:       accessKey,
		SecretKey:       secretKey,
		Bucket:          readString("STORAGE_BUCKET", "street-born-assets"),
		Region:          readString("STORAGE_REGION", "us-east-1"),
		UseSSL:          useSSL,
		PresignLifetime: presignLifetime,
	}, nil
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
		FromAddress:  readString("EMAIL_FROM_ADDRESS", "no-reply@streetborn.test"),
		FromName:     readString("EMAIL_FROM_NAME", "Street Born"),
	}, nil
}

func loadSessionConfig(environment string) (SessionConfig, error) {
	jwtSecret := readString("JWT_SECRET", "")
	if len(jwtSecret) < MinimumJWTSecretLength {
		return SessionConfig{}, fmt.Errorf("JWT_SECRET must be at least %d characters", MinimumJWTSecretLength)
	}
	isCookieSecure, err := readBool("COOKIE_SECURE", environment != EnvironmentDevelopment)
	if err != nil {
		return SessionConfig{}, err
	}
	return SessionConfig{JWTSecret: jwtSecret, IsCookieSecure: isCookieSecure}, nil
}

func LoadDatabaseConfig() (DatabaseConfig, error) {
	databaseURL := readString("DATABASE_URL", "")
	if databaseURL == "" {
		return DatabaseConfig{}, errors.New("DATABASE_URL is required")
	}

	maxOpenConnections, err := readInt("DATABASE_MAX_CONNECTIONS", 20)
	if err != nil {
		return DatabaseConfig{}, err
	}
	if maxOpenConnections < 1 {
		return DatabaseConfig{}, errors.New("DATABASE_MAX_CONNECTIONS must be at least 1")
	}

	connectionMaxLifetime, err := readDuration("DATABASE_CONNECTION_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return DatabaseConfig{}, err
	}

	return DatabaseConfig{
		URL:                   databaseURL,
		MaxOpenConnections:    maxOpenConnections,
		MaxIdleConnections:    max(maxOpenConnections/4, 1),
		ConnectionMaxLifetime: connectionMaxLifetime,
	}, nil
}

func (cfg Config) IsProduction() bool {
	return cfg.Environment == EnvironmentProduction
}

func (cfg Config) HTTPAddress() string {
	return fmt.Sprintf(":%d", cfg.HTTPPort)
}

func loadEnvironment() (string, error) {
	environment := readString("APP_ENV", EnvironmentDevelopment)
	switch environment {
	case EnvironmentDevelopment, EnvironmentProduction, EnvironmentTest:
		return environment, nil
	default:
		return "", fmt.Errorf("APP_ENV %q is not one of development, production, test", environment)
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

func loadWalletPaymentConfig() (WalletPaymentConfig, error) {
	requiredConfirmations, err := readInt("CHAIN_REQUIRED_CONFIRMATIONS", defaultRequiredConfirmations)
	if err != nil {
		return WalletPaymentConfig{}, err
	}
	if requiredConfirmations < 1 {
		return WalletPaymentConfig{}, errors.New("CHAIN_REQUIRED_CONFIRMATIONS must be 1 or more")
	}
	pollInterval, err := readDuration("CHAIN_POLL_INTERVAL", defaultChainPollInterval)
	if err != nil {
		return WalletPaymentConfig{}, err
	}
	return WalletPaymentConfig{
		RPCURL:                     readString("CHAIN_RPC_URL", ""),
		ChapterPaymentVaultAddress: readString("CHAPTER_PAYMENT_VAULT_ADDRESS", ""),
		PaymentSignerPrivateKey:    readString("PAYMENT_SIGNER_PRIVATE_KEY", ""),
		RequiredConfirmations:      requiredConfirmations,
		PollInterval:               pollInterval,
	}, nil
}
