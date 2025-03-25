/*
vxinst - Blazing fast embedder for instagram posts
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
	"bitwise7/vxinst/api/public"
	"bitwise7/vxinst/flags"
	"bitwise7/vxinst/utils"
	"log/slog"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	h := public.NewHandler(db)
	h.Init()

	// Initialize ticker for database cleanup
	go cleanDb(db)

	if *flags.Secure {
		slog.Info("Server running with TLS enabled", slog.String("listen", *flags.Port))
		h.Router.RunTLS(":"+*flags.Port, *flags.CertFile, *flags.KeyFile)
	} else {
		slog.Info("Server running", slog.String("listen", *flags.Port))
		h.Router.Run(":" + *flags.Port)
	}

}

// Periodically cleans up expired records from the database
func cleanDb(db *gorm.DB) {
	ticker := time.NewTicker(5 * time.Minute)

	for range ticker.C {
		slog.Debug("Tick! Cleaning up records")

		toDelete := []string{}

		err := db.
			Model(&utils.HtmlData{}).
			Where("expires_at < ?", time.Now().Unix()).
			Select("shortcode").
			Find(&toDelete).
			Error

		if err != nil {
			if err == gorm.ErrRecordNotFound {
				slog.Debug("No records to cleanup")
				return
			}

			slog.Error("Failed to cleanup records", slog.Any("err", err))
			return
		}

		tx := db.Begin()
		tx.Model(&utils.HtmlData{}).Where("shortcode IN ?", toDelete).Delete(nil)

		if err := tx.Commit().Error; err != nil {
			slog.Error("Failed to commit database transaction", slog.Any("err", err.Error))
			return
		} else {
			slog.Debug("Old records deleted")
		}
	}
}
