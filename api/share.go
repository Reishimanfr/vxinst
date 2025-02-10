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
	"net/http"

	"github.com/gin-gonic/gin"
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
func FollowShare(c *gin.Context) {
	// span := sentry.StartSpan(c.Request.Context(), "share.parse")
	// defer span.Finish()

	// req, err := http.NewRequest("GET", "https://instagram.com"+c.Request.URL.String(), nil)
	// if err != nil {
	// 	slog.Error("Failed to prepare request to follow redirects", slog.Any("err", err))
	// 	sentry.CaptureException(err)
	// 	c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
	// 		Title:       "VxInstagram - Server Error",
	// 		Description: "VxInstagram encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
	// 	})
	// 	return
	// }

	// res, err := client.Do(req)
	// if err != nil {
	// 	slog.Error("Failed to follow redirects", slog.Any("err", err))
	// 	sentry.CaptureException(err)
	// 	c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
	// 		Title:       "VxInstagram - Server Error",
	// 		Description: "VxInstagram encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
	// 	})
	// 	return
	// }
	// res.Body.Close()

	// urlSplit := strings.Split(res.Request.URL.String(), "/")
	// postId := urlSplit[len(urlSplit)-2]

	// userAgent := strings.ToLower(c.Request.Header.Get("User-Agent"))

	// // Redirect browsers to the post
	// if !strings.Contains(userAgent, "discord") {
	// 	c.Redirect(http.StatusPermanentRedirect, "https://instagram.com/reel/"+postId)
	// 	return
	// }

	// videoUrl, err := utils.ParseGQLData(postId)
	// if err != nil {
	// 	slog.Error("Failed to get video URL from instagram's CDN", slog.Any("err", err))
	// 	sentry.CaptureException(err)
	// 	c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
	// 		Title:       "VxInstagram - Server Error",
	// 		Description: "VxInstagram encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
	// 	})
	// 	return
	// }

	// if videoUrl == "" {
	// 	sentry.CaptureMessage("Instagram returned an empty video URL.")
	// 	c.HTML(http.StatusNoContent, "embed.html", &HtmlOpenGraphData{
	// 		Title:       "VxInstagram - Empty Response",
	// 		Description: "Instagram returned an empty URL. You'll need to watch this post in your browser. Sorry!",
	// 	})
	// 	return
	// }

	// remote, err := url.Parse(videoUrl)
	// if err != nil {
	// 	slog.Error("Failed to parse CDN video URL", slog.Any("err", err))
	// 	sentry.CaptureException(err)
	// 	c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
	// 		Title:       "VxInstagram - Server Error",
	// 		Description: "VxInstagram encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
	// 	})
	// 	return
	// }

	// proxy := httputil.NewSingleHostReverseProxy(remote)
	// proxy.Director = func(r *http.Request) {
	// 	r.Header = c.Request.Header
	// 	r.Host = remote.Host
	// 	r.URL = remote
	// 	r.Header = c.Request.Header.Clone()

	// 	hopHeaders := []string{
	// 		"Connection", "Keep-Alive", "Proxy-Authenticate", "Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding",
	// 	}

	// 	for _, h := range hopHeaders {
	// 		r.Header.Del(h)
	// 	}
	// }

	// c.Header("Cache-Control", "max-age=43200")
	// proxy.ServeHTTP(c.Writer, c.Request)
}
