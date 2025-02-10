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
	"bash06/vxinstagram/utils"
	"context"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

var ctx = context.Background()

type HtmlOpenGraphData struct {
	Title       string
	Description string
	URL         string
	VideoURL    string
	Color       string
}

func ServeVideo(c *gin.Context) {
	postId := c.Param("id")

	span := sentry.StartSpan(c.Request.Context(), "serve.video")
	defer span.Finish()

	if postId == "" || postId[0] != 'D' && postId[0] != 'C' {
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "VxInstagram - Not found",
			Description: "An invalid post ID was provided. Please make sure the URL is correct",
		})
		return
	}

	userAgent := strings.ToLower(c.Request.Header.Get("User-Agent"))

	// Redirect browsers to the post
	if !strings.Contains(userAgent, "discord") {
		c.Redirect(http.StatusPermanentRedirect, "https://instagram.com/reel/"+postId)
		return
	}

	// videoUrl, err := utils.GetCdnUrl(postId)
	// if err != nil || videoUrl == "" {
	// 	slog.Warn("Failed to scrape video URL from HTML. Trying with GraphQL")

	// 	videoUrl, err = utils.ParseGQLData(postId)
	// }

	var videoUrl string
	var err error

	videoUrl, err = utils.GetCdnUrl(postId)
	if err != nil {
		slog.Error("Failed to get CDN video URL", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "VxInstagram - Not found",
			Description: "An invalid post ID was provided. Please make sure the URL is correct",
		})
		return
	}
	// if err != nil || videoUrl == "" {
	// 	slog.Warn("Failed to scrape video URL from HTML. Trying to scrape from GraphQL")

	// 	attempt := 0

	// 	for {
	// 		if attempt >= 5 {
	// 			slog.Error("Failed to scrape video URL after 5 attempts")
	// 			break
	// 		}

	// 		attempt++

	// 		slog.Debug("Trying to get video url", slog.Int("attempt", attempt))

	// 		videoUrl, err = utils.GetCdnUrl(postId)
	// 		if err != nil {
	// 			slog.Debug("Failed to get video url", slog.Int("attempt", attempt))
	// 			continue
	// 		}

	// 		if videoUrl == "" {
	// 			slog.Debug("Got an empty URL, trying again...")
	// 			continue
	// 		}

	// 		if videoUrl != "" {
	// 			break
	// 		}
	// 	}
	// }

	if videoUrl == "" {
		slog.Warn("Instagram returned an empty video URL")
		sentry.CaptureMessage("Instagram returned an empty video URL")
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "VxInstagram - Empty Response",
			Description: "Instagram returned an empty URL. You'll need to watch this post in your browser. Sorry!",
		})
		return
	}

	remote, err := url.Parse(videoUrl)
	if err != nil {
		slog.Error("Failed to parse CDN video URL", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "VxInstagram - Server Error",
			Description: "VxInstagram encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
		})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Director = func(r *http.Request) {
		r.Header = c.Request.Header
		r.Host = remote.Host
		r.URL = remote
		r.Header = c.Request.Header.Clone()

		hopHeaders := []string{
			"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding",
		}

		for _, h := range hopHeaders {
			r.Header.Del(h)
		}
	}

	c.Header("Cache-Control", "max-age=43200")
	proxy.ServeHTTP(c.Writer, c.Request)

	// c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
	// 	Title:       "VxInstagram",
	// 	Description: "",
	// 	URL:         "https://instagram.com/reel/" + postId,
	// 	VideoURL:    videoUrl,
	// })
}
