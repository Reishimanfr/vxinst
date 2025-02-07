package api

import (
	"bash06/vxinstagram/utils"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func ServeVideo(c *gin.Context) {
	postId := c.Param("id")

	span := sentry.StartSpan(c.Request.Context(), "serve.video")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"Origin": "https://instagram.com/reel/" + postId,
	}

	if postId == "" || postId[0] != 'D' {
		c.HTML(http.StatusNotFound, "invalid_id.html", nil)
		return
	}

	userAgent := strings.ToLower(c.Request.Header.Get("User-Agent"))

	// Redirect browsers to the post
	if !strings.Contains(userAgent, "discord") {
		span.Data["Redirect"] = true
		c.Redirect(http.StatusPermanentRedirect, "https://instagram.com/reel/"+postId)
		return
	}

	videoUrl, err := utils.GetCdnUrl(postId)
	if err != nil {
		slog.Error("Failed to get video URL from instagram's CDN", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusOK, "server_error.html", nil)
		return
	}

	if videoUrl == "" {
		slog.Warn("Instagram returned an empty video URL. This most likely means the video is age restricted")
		sentry.CaptureMessage("Instagram returned an empty video URL. This most likely means the video is age restricted")
		c.HTML(http.StatusOK, "no_url.html", nil)
		return
	}

	remote, err := url.Parse(videoUrl)
	if err != nil {
		slog.Error("Failed to parse CDN video URL", slog.Any("err", err))
		c.HTML(http.StatusOK, "server_error.html", nil)
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
}
