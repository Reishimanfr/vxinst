package main

import (
	"bash06/vxinstagram/api"
	"bash06/vxinstagram/flags"
	"log/slog"
	"net/http"
	"os"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
)

var store = persist.NewMemoryStore(time.Minute * 1)

func main() {
	if err := flags.Parse(); err != nil {
		panic(err)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              *flags.SentryDsn,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	}); err != nil {
		slog.Error("Failed to initialize sentry", slog.Any("err", err))
	}

	var level slog.Level

	switch *flags.LogLevel {
	case "error":
		level = slog.LevelError
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "debug":
		level = slog.LevelDebug
	default: // Impossible case
		level = slog.LevelInfo
	}

	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      level,
			TimeFormat: time.Kitchen,
		}),
	))

	// if *flags.SentryDsn != "" {
	// fmt.Println("initializing sentryr")

	// }

	defer sentry.Flush(time.Second * 2)

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.ErrorLogger())
	r.Use(api.RateLimiterMiddleware(api.NewRateLimiter(5, 10)))
	r.Use(sentrygin.New(sentrygin.Options{}))
	r.LoadHTMLGlob("templates/*")

	if *flags.GinLogs {
		r.Use(gin.Logger())
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r.GET("/reel/:id", cache.CacheByRequestURI(store, time.Minute*1), api.ServeVideo)
	r.GET("/reels/:id", cache.CacheByRequestURI(store, time.Minute*1), api.ServeVideo)
	r.GET("/p/:id", cache.CacheByRequestURI(store, time.Minute*1), api.ServeVideo)
	r.GET("/favicon.ico1", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	r.GET("/share/:id", cache.CacheByRequestURI(store, time.Minute*1), api.FollowShare)

	if *flags.Secure {
		slog.Info("Server running with TLS enabled", slog.String("listen", *flags.Port))
		r.RunTLS(":"+*flags.Port, *flags.CertFile, *flags.KeyFile)
	} else {
		slog.Info("Server running", slog.String("listen", *flags.Port))
		r.Run(":" + *flags.Port)
	}
}
