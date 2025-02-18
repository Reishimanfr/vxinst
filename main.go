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
	"bash06/vxinstagram/utils"
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
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

	db, err := utils.InitDb()
	if err != nil {
		slog.Error("Failed to initialize database", slog.Any("err", err))
		os.Exit(1)
	}

	h := api.NewHandler(db)
	h.Init()

	if *flags.Secure {
		slog.Info("Server running with TLS enabled", slog.String("listen", *flags.Port))
		h.Router.RunTLS(":"+*flags.Port, *flags.CertFile, *flags.KeyFile)
	} else {
		slog.Info("Server running", slog.String("listen", *flags.Port))
		h.Router.Run(":" + *flags.Port)
	}
}
