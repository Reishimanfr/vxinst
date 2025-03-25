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
	"log/slog"
	"net/http"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

var (
	client = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
)

// Videos shared from the phone generate a redirect
// ID, then redirect the user to the actual post.
// This means we have to first follow the redirects before we can actually embed
// the post itself which is extremely annoying and slow.
// Lots of fuckery and workarounds just to support one edge case.
func (h *Handler) FollowShare(c *gin.Context) {
	span := sentry.StartSpan(c.Request.Context(), "share.parse")
	defer span.Finish()

	req, err := http.NewRequest("GET", "https://instagram.com"+c.Request.URL.String(), nil)
	if err != nil {
		slog.Error("Failed to prepare request to follow redirects", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "vxinst - Server Error",
			Description: "vxinst encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
		})
		return
	}

	res, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to follow redirects", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusOK, "embed.html", &HtmlOpenGraphData{
			Title:       "vxinst - Server Error",
			Description: "vxinst encountered a server side error while processing your request. Request ID:`" + span.SpanID.String() + "`",
		})
		return
	}
	res.Body.Close()

	urlSplit := strings.Split(res.Request.URL.String(), "/")
	postId := urlSplit[len(urlSplit)-2]

	h.ProcessPost(c, postId)
}
