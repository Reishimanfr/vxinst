package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gin-gonic/gin"
	"github.com/lmittmann/tint"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type VideoCache struct {
	Origin   string `gorm:"primaryKey"`
	Dest     string
	Lifetime int64
}

var (
	db  *gorm.DB
	err error
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelInfo,
			TimeFormat: time.Kitchen,
		}),
	))

	db, err = gorm.Open(sqlite.Open("data.db"))
	if err != nil {
		panic(fmt.Errorf("failed to initialize sqlite database: %v", err))
	}

	err = db.AutoMigrate(&VideoCache{})
	if err != nil {
		panic(fmt.Errorf("failed to automigrate tables: %v", err))
	}

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(gin.ErrorLogger())
	gin.SetMode(gin.ReleaseMode)

	r.GET("/reel/:id", ServeReel)
	r.GET("/p/:id", ServeReel)
	r.Run()
}

func ServeReel(c *gin.Context) {
	postId := c.Param("id")

	url := "https://instagram.com/p/" + postId + "/embed/captioned"

	var cache *VideoCache
	err := db.
		Where("origin = ?", url).
		First(&cache).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			slog.Info("Fetching video that's not found in cache")
		} else {
			slog.Error("Failed to read video cache. Trying to scrape the video instead...", slog.Any("err", err))
		}
	}

	// Cache is invalid
	if cache.Lifetime < time.Now().Unix() {
		cache = &VideoCache{}
	}

	if cache.Dest == "" {
		ctx, cancel := chromedp.NewContext(context.Background())
		defer cancel()

		var videoUrl string

		err := chromedp.Run(ctx,
			chromedp.Navigate(url),
			chromedp.WaitVisible("video", chromedp.ByQuery),
			chromedp.Evaluate(`document.querySelector("video").src`, &videoUrl),
		)
		if err != nil {
			slog.Error("Failed to scrape video src", slog.Any("err", err))

			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		cache = &VideoCache{
			Origin:   url,
			Dest:     videoUrl,
			Lifetime: time.Now().Add(time.Minute * 15).Unix(),
		}
	}

	err = db.
		Where("origin = ?", url).
		Save(&cache).
		Error
	if err != nil {
		slog.Error("Failed to save video cache", slog.Any("err", err))
	}

	c.Header("Content-Type", "video/mp4")
	c.Redirect(http.StatusTemporaryRedirect, cache.Dest)
}
