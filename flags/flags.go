package flags

import (
	"flag"
	"fmt"
	"slices"
	"strconv"
)

var (
	Port      = flag.String("port", "8080", "Port to run the server on")
	GinLogs   = flag.Bool("gin-logs", false, "Enable gin debug logs")
	Secure    = flag.Bool("secure", false, "Use a secure connection")
	LogLevel  = flag.String("log-level", "info", "Logging verbositily level [debug, error, warn, info]")
	CertFile  = flag.String("cert-file", "", "Path to the SSL certificate (only needed with secure enabled)")
	KeyFile   = flag.String("key-file", "", "Path to the SSL key (only needed with secure enabled)")
	SentryDsn = flag.String("sentry-dsn", "", "Sentry DSN used for telementry")
)

func Parse() error {
	flag.Parse()

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
