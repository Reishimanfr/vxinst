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
package main 

import (
	"bash06/vxinstagram/api"
	"bash06/vxinstagram/flags"
	"log/slog"
	"net/http"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func main() {
	flags.Parse()

	if !*flags.GinLogs {
		gin.SetMode(gin.ReleaseMode)
	}

	// Don't try to initialize sentry if no DSN provided
	if *flags.SentryDsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              *flags.SentryDsn,
			EnableTracing:    true,
			TracesSampleRate: 1.0,
		}); err != nil {
			slog.Error("Failed to initialize sentry", slog.Any("err", err))
		}
	}
	defer sentry.Flush(time.Second * 2)

	cacheExpire := time.Minute * time.Duration(*flags.CacheLifetime)
	var st persist.CacheStore = persist.NewMemoryStore(cacheExpire)

	if *flags.RedisEnable {
		rdb := redis.NewClient(&redis.Options{
			Addr:     *flags.RedisAddr,
			Password: *flags.RedisPasswd,
			DB:       *flags.RedisDB,
		}) 
		st = persist.NewRedisStore(rdb)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.ErrorLogger())
	r.Use(api.RateLimiterMiddleware(api.NewRateLimiter(5, 10)))
	r.Use(sentrygin.New(sentrygin.Options{}))
	r.LoadHTMLGlob("templates/*")

	// Endpoints
	r.GET("/reel/:id", cache.CacheByRequestURI(st, cacheExpire), api.ServeVideo)
	r.GET("/reels/:id", cache.CacheByRequestURI(st, cacheExpire), api.ServeVideo)
	r.GET("/p/:id", cache.CacheByRequestURI(st, cacheExpire), api.ServeVideo)
	r.GET("/favicon.ico", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })

	r.GET("/share/:id", cache.CacheByRequestURI(st, cacheExpire), api.FollowShare)

	// Redirect vxinstagram.com to README
	r.GET("/", func(ctx *gin.Context) {
		ctx.Redirect(http.StatusPermanentRedirect, "https://github.com/Reishimanfr/vxinstagram?tab=readme-ov-file#how-to-use")
	})

	if *flags.Secure {
		slog.Info("Server running with TLS enabled", slog.String("listen", *flags.Port))
		r.RunTLS(":"+*flags.Port, *flags.CertFile, *flags.KeyFile)
	} else {
		slog.Info("Server running", slog.String("listen", *flags.Port))
		r.Run(":" + *flags.Port)
	}
}
