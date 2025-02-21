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
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
)

// Videos shared from the phone work in a different way. They generate a redirect
// ID, then redirect the user to the actual post.
// This means we have to first follow the redirects before we can actually embed
// the post itself which is extremely annoying and slow
func (h *Handler) FollowShare(c *gin.Context) {
	span := sentry.StartSpan(c.Request.Context(), "share.parse")
	defer span.Finish()

	req, err := http.NewRequest("GET", "https://instagram.com"+c.Request.URL.String(), nil)
	if err != nil {
		slog.Error("Failed to prepare request to follow redirects", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "VxInstagram - Server Error",
			Description: "VxInstagram encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
		})
		return
	}

	res, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to follow redirects", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "VxInstagram - Server Error",
			Description: "VxInstagram encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
		})
		return
	}
	res.Body.Close()

	urlSplit := strings.Split(res.Request.URL.String(), "/")
	postId := urlSplit[len(urlSplit)-2]

	if *flags.RedirectBrowsers {
		slog.Debug("Redirecting browser to instagram post")
		userAgent := strings.ToLower(c.Request.Header.Get("User-Agent"))

		if !strings.Contains(userAgent, "discord") {
			c.Redirect(http.StatusPermanentRedirect, "https://instagram.com/reel/"+postId)
			return
		}
	}

	var record *utils.ExtractedData
	var data *utils.ExtractedData
	var videoUrl string
	create := false

	err = h.Db.
		Model(utils.Post{}).
		Where("id = ?", postId).
		Select("title", "post_url", "video_url", "thumbnail_url").
		First(&record).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			slog.Debug("Database doesn't include CDN record")
			create = true
		} else {
			slog.Error("Failed to read cache from database", slog.Any("err", err))
		}

		// 1: Try to scrape the HTML
		data, err = utils.ScrapeFromHTML(postId)
		videoUrl = data.VideoURL

		if err != nil || videoUrl == "" {
			slog.Error("Failed to scrape video URL from HTML. Trying to make an API request...", slog.Any("err", err))

			if err != nil {
				sentry.CaptureException(err)
			}

			// 2: Try to get the post data from an API request
			data, err := utils.FetchPost(postId)
			if err != nil {
				slog.Error("Failed to fetch video URL from API. Critial failure!", slog.Any("err", err))
				sentry.CaptureException(err)

				if err.Error() != "no instagram cookie provided" {
					c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
						Title:       "VxInstagram - Server Error",
						Description: "VxInstagram encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
					})
				}
			} else {
				videoUrl = data.Items[0].VideoVersions[0].URL
			}
		}
	}

	if videoUrl == "" && data.ThumbnailURL != "" {
		slog.Debug("Post didn't have a video but we found an image to show")

		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:    "@" + data.Username,
			ImageURL: data.ThumbnailURL,
		})
	}

	if videoUrl == "" {
		slog.Debug("No video URL found! :(")
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "VxInstagram - Empty Response",
			Description: "Instagram returned an empty response meaning we can't embed the video. You'll need to watch it in your browser. Sorry!",
		})
	}

	c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
		Title:    "@" + data.Username,
		VideoURL: videoUrl,
		PostURL:  "https://instagram.com/" + c.Request.URL.String(),
	})

	slog.Debug("Done!")

	if create {
		err := h.Db.Model(&utils.Post{}).Create(&utils.Post{
			Id:           postId,
			VideoURL:     videoUrl,
			ThumbnailURL: data.ThumbnailURL,
			Title:        "@" + data.Username,
			PostURL:      "https://instagram.com/" + c.Request.URL.String(),
			ExpiresAt:    time.Now().Add(time.Hour * 168).Unix(), // TODO: make this a flag
		}).Error
		if err != nil {
			sentry.CaptureException(err)
			slog.Error("Failed to save cdn url to memory database", slog.Any("err", err))
		}
	}
}
