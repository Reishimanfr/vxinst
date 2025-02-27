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
package api

import (
	"bash06/vxinstagram/flags"
	"bash06/vxinstagram/utils"
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	ctx            = context.Background()
	memoryLifetime = time.Now().Add(time.Hour * time.Duration(24*(*flags.MemoryLifetime))).Unix()
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

	span := sentry.StartSpan(c.Request.Context(), "serve.video")
	defer span.Finish()

	if postId == "" || postId[0] != 'D' && postId[0] != 'C' {
		slog.Debug("Invalid post id provided")
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "VxInstagram - Not found",
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

	var data *utils.ExtractedData
	var scrapeErr error
	create := false

	if err := h.Db.Model(&utils.ExtractedData{}).Where("id = ?", postId).First(&data).Error; err != nil {
		create = true

		if err == gorm.ErrRecordNotFound {
			slog.Debug("No record found. Fetching new data")
		} else {
			slog.Error("Failed to read cache from database", slog.Any("err", err))
		}

		// 1: Try to scrape from HTML
		data, scrapeErr = utils.ScrapeFromHTML(postId)
		if scrapeErr != nil {
			slog.Error("Failed to scrape from HTML", slog.Any("err", err))
			sentry.CaptureException(err)
		} else if data == nil {
			slog.Debug("No video URL found in HTML. Trying to fetch from API")

			igResp, err := utils.FetchPost(postId)
			if err != nil && err.Error()[0:8] != "bad flag" {
				slog.Error("Failed to fetch data from API", slog.Any("err", err))
				sentry.CaptureException(err)
			} else if igResp != nil && len(igResp.Items) > 0 && len(igResp.Items[0].VideoVersions) > 0 && len(igResp.Items[0].ImageVersions.Candidates) > 0 {
				data = &utils.ExtractedData{
					VideoURL:     igResp.Items[0].VideoVersions[0].URL,
					ThumbnailURL: igResp.Items[0].ImageVersions.Candidates[0].URL,
					Username:     "",
				}
			}
		}
	} else {
		slog.Debug("Found record in database")
	}

	postURL := "https://instagram.com/" + c.Request.URL.String()

	if create {
		slog.Debug("Creating new record in database")

		newRecord := &utils.ExtractedData{
			Id:        postId,
			ExpiresAt: memoryLifetime,
		}

		if data != nil {
			newRecord.VideoURL = data.VideoURL
			newRecord.ThumbnailURL = data.ThumbnailURL
			newRecord.Username = data.Username
			newRecord.Views = data.Views
			newRecord.Comments = data.Comments
			newRecord.Likes = data.Likes
			newRecord.PostURL = postURL
		}

		if err := h.Db.Model(&utils.ExtractedData{}).Create(newRecord).Error; err != nil {
			sentry.CaptureException(err)
			slog.Error("Failed to save record to memory database", slog.Any("err", err))
		}
	}

	// Case 1: No data at all
	if data == nil {
		slog.Debug("No data found in database or from scraping")
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "VxInstagram - Empty Response",
			Description: "Instagram returned an empty response meaning we can't embed the post. You'll need to see it in your browser. Sorry!",
		})
		return
	}

	// Case 1: No video URL found, but we have a thumbnail
	if data.VideoURL == "" && data.ThumbnailURL != "" {
		slog.Debug("Post didn't have a video but we found an image to show")

		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:    "@" + data.Username,
			ImageURL: data.ThumbnailURL,
		})
		return
	}

	// Case 3: We have a video URL
	c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
		Title:       "Post by @" + data.Username,
		Description: "L: " + data.Likes + " C: " + data.Comments + " V: " + data.Views,
		VideoURL:    data.VideoURL,
		PostURL:     postURL,
	})
}
