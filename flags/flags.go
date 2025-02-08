package flags

import (
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/spf13/pflag"
)

var (
	Port      = pflag.StringP("port", "p", getEnvDefault("PORT", "8080"), "Port to run the server on")
	GinLogs   = pflag.BoolP("gin-logs", "g", getEnvDefaultBool("GIN_LOGS", false), "Enable gin debug logs")
	Secure    = pflag.BoolP("secure", "s", getEnvDefaultBool("SECURE", false), "Use a secure connection")
	LogLevel  = pflag.StringP("log-level", "l", getEnvDefault("LOG_LEVEL", "info"), "Logging verbositily level [debug, error, warn, info]")
	CertFile  = pflag.StringP("cert-file", "c", getEnvDefault("CERT_FILE", ""), "Path to the SSL certificate (only needed with secure enabled)")
	KeyFile   = pflag.StringP("key-file", "k", getEnvDefault("KEY_FILE", ""), "Path to the SSL key (only needed with secure enabled)")
	SentryDsn = pflag.StringP("sentry-dsn", "d", getEnvDefault("SENTRY_DSN", ""), "Sentry DSN used for telemetry")
)

func getEnvDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvDefaultBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return defaultValue
}

func Parse() error {
	pflag.Parse()

	if _, err := strconv.Atoi(*Port); err != nil {
		return fmt.Errorf("port is not a valid integer")
	}

	if *Secure {
		if CertFile == nil {
			return fmt.Errorf("no ssl certificate file provided")
		}

		if KeyFile == nil {
			return fmt.Errorf("no ssl key file provided")
		}
	}

	if !slices.Contains([]string{"debug", "error", "warn", "info"}, *LogLevel) {
		return fmt.Errorf("invalid logging level provided: \"%v\". Expected one of [debug, error, warn, info]", *LogLevel)
	}

	return nil
}
