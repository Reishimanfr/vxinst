/*
VxInstagram - Blazing fast embedder for instagram posts
Copyright (C) 2025 Bash06

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package flags

import (
	"io/fs"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/pflag"
)

var (
	Port      = pflag.StringP("port", "p", getEnvDefault("PORT", "8080"), "Port to run the server on")
	GinLogs   = pflag.BoolP("gin-logs", "g", getEnvDefaultBool("GIN_LOGS", false), "Enable gin debug logs")
	Secure    = pflag.BoolP("secure", "s", getEnvDefaultBool("SECURE", false), "Use a secure connection")
	LogLevel  = pflag.StringP("log-level", "v", getEnvDefault("LOG_LEVEL", "info"), "Logging verbosity level [debug, error, warn, info]")
	CertFile  = pflag.StringP("cert-file", "C", getEnvDefault("CERT_FILE", ""), "Path to the SSL certificate (only needed with secure enabled)")
	KeyFile   = pflag.StringP("key-file", "K", getEnvDefault("KEY_FILE", ""), "Path to the SSL key (only needed with secure enabled)")
	SentryDsn = pflag.StringP("sentry-dsn", "d", getEnvDefault("SENTRY_DSN", ""), "Sentry DSN used for telemetry")

	CacheLifetime = pflag.IntP("cache-lifetime", "L", getEnvDefaultInt("CACHE_LIFETIME", 60), "Cache lifetime (in minutes)")

	RedisEnable = pflag.BoolP("redis-enable", "r", getEnvDefaultBool("REDIS_ENABLE", false), "Enables redis")
	RedisAddr   = pflag.StringP("redis-address", "A", getEnvDefault("REDIS_ADDR", ""), "Address to redis database for caching")
	RedisPasswd = pflag.StringP("redis-passwd", "P", getEnvDefault("REDIS_PASSWD", ""), "Password to redis database")
	RedisDB     = pflag.IntP("redis-db", "D", getEnvDefaultInt("REDIS_DB", -1), "Redis database to use")

	logLevels = []string{"debug", "info", "warn", "error"}
)

func getEnvDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvDefaultInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
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

func Parse() {
	pflag.Parse()

	var level slog.Level

	switch *LogLevel {
	case "error":
		level = slog.LevelError
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "debug":
		level = slog.LevelDebug
	default:
		level = slog.LevelInfo
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
			AddSource:  true,
		}),
	))

	if !slices.Contains(logLevels, *LogLevel) {
		slog.Warn("Invalid logging level provided. Falling back to 'info'", slog.String("level", *LogLevel))
	}

	if _, err := strconv.Atoi(*Port); err != nil {
		slog.Error("Port is not a valid integer", slog.String("port", *Port))
		os.Exit(1)
	}

	if *CacheLifetime <= 0 {
		slog.Error("Cache lifetime must be greater than 0", slog.Int("lifetime", *CacheLifetime))
		os.Exit(1)
	}

	if *RedisEnable && *RedisDB == -1 {
		slog.Error("No redis database provided")
		os.Exit(1)
	}

	if *Secure {
		if *CertFile == "" {
			slog.Error("No SSL certificate file path provided")
			os.Exit(1)
		}

		if file, err := os.Stat(*CertFile); err != nil {
			if err == fs.ErrPermission {
				slog.Warn("Insufficient permissions to check if certificate file exists")
			}

			slog.Error("Certificate file at path doesn't exist", slog.String("path", *CertFile))
			os.Exit(1)
		} else {
			if file.IsDir() {
				slog.Error("Certificate file at path doesn't exist", slog.String("path", *CertFile))
				os.Exit(1)
			}
		}

		if *KeyFile == "" {
			slog.Error("No SSL key file path provided")
			os.Exit(1)
		}

		if file, err := os.Stat(*KeyFile); err != nil {
			if err == fs.ErrPermission {
				slog.Warn("Insufficient permissions to check if certificate key file exists")
			}

			slog.Error("Certificate key file at path doesn't exist", slog.String("path", *KeyFile))
			os.Exit(1)
		} else {
			if file.IsDir() {
				slog.Error("Certificate key file at path doesn't exist", slog.String("path", *KeyFile))
				os.Exit(1)
			}
		}
	}
}
