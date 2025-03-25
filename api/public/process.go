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
package public

import (
	"bitwise7/vxinst/flags"
	"bitwise7/vxinst/utils"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ctx                  = context.Background()
	scrapingMethodsFuncs = map[string]func(postId string) (*utils.HtmlData, error){
		"html": utils.ScrapeFromHTML,
		// "graphql": utils.ScrapeFromGQL,
		// "api":     utils.FetchPost,
	}
)

type HtmlOpenGraphData struct {
	Title       string
	Description string
	VideoURL    string
	ImageURL    string
	PostURL     string
}

// Shared portion between some endpoints that do the same thing with minor
// differences. Post ID must be specified since it's returned in different ways for each endpoint
func (h *Handler) ProcessPost(c *gin.Context, postId string) {
	slog.Debug("Got a request to process post", slog.String("id", postId))

	if postId == "" || postId[0] != 'D' && postId[0] != 'C' {
		slog.Debug("Invalid post id provided")
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "vxinst - Not found",
			Description: "An invalid post ID was provided. Please make sure the URL is correct",
		})
		return
	}

	if *flags.RedirectBrowsers {
		userAgent := strings.ToLower(c.Request.Header.Get("User-Agent"))

		if !strings.Contains(userAgent, "discord") {
			slog.Debug("Redirecting browser to instagram post")
			c.Redirect(http.StatusPermanentRedirect, "https://instagram.com/"+c.Request.URL.String())
			return
		}
	}

	create := false
	var data *utils.HtmlData
	if err := h.Db.Model(&utils.HtmlData{}).Where("shortcode = ?", postId).First(&data).Error; err != nil {
		create = true

		if err == gorm.ErrRecordNotFound {
			slog.Debug("No record found. Fetching new data")
		} else {
			slog.Error("Failed to read cache from database", slog.Any("err", err))
		}

		for _, method := range *flags.ScrapingMethods {
			fn, ok := scrapingMethodsFuncs[method]

			if !ok {
				slog.Debug("Invalid scraping method", slog.String("method", method))
				continue
			}

			slog.Debug("Trying method", slog.String("method", method))

			data, err = fn(postId)
			if err != nil {
				slog.Error("Method failed, trying something else if available", slog.Any("err", err))
				continue
			}

			if data == nil {
				slog.Error("Method didn't get any data, trying something else if available")
				continue
			} else {
				slog.Debug("Found some data")
			}

			fmt.Println(data)

			break
		}
	} else {
		slog.Debug("Found record in database")
	}

	if create {
		slog.Debug("Creating new record in database")

		newRecord := &utils.HtmlData{
			Shortcode: postId,
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(24*(*flags.MemoryLifetime))).Unix(),
		}

		if data != nil {
			newRecord = data
			newRecord.ExpiresAt = time.Now().Add(time.Hour * time.Duration(24*(*flags.MemoryLifetime))).Unix()
		}

		if err := h.Db.Model(&utils.HtmlData{}).Create(newRecord).Error; err != nil {
			sentry.CaptureException(err)
			slog.Error("Failed to save record to memory database", slog.Any("err", err))
		}
	}

	// Case 1: No data at all
	if data == nil {
		slog.Debug("No data found in database or from scraping")
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "vxinst - Empty Response",
			Description: "Instagram returned an empty response meaning we can't embed the post. You'll need to see it in your browser. Sorry!",
		})
		return
	}

	// Case 1: No video URL found, but we have a thumbnail
	if data.Video.URL == "" && data.ThumbnailURL != "" {
		slog.Debug("Post didn't have a video but we found an image to show")

		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:    "@" + data.Author.Username,
			ImageURL: data.ThumbnailURL,
		})
		return
	}

	var sb strings.Builder

	sb.WriteString("❤️: ")
	sb.WriteString(strconv.Itoa(data.Likes))
	sb.WriteString(" 💬: ")
	sb.WriteString(strconv.Itoa(data.Comments))
	sb.WriteString(" 👁️: ")
	sb.WriteString(strconv.Itoa(data.Views))

	// Case 3: We have a video URL
	c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
		Title:       "Post by @" + data.Author.Username,
		Description: sb.String(),
		VideoURL:    data.Video.URL,
		PostURL:     data.Permalink,
	})
}
