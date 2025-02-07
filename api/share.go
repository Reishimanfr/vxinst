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
	span := sentry.StartSpan(c.Request.Context(), "serve.video")
	defer span.Finish()

	span.Data = map[string]interface{}{
		"Origin": "https://instagram.com" + c.Request.URL.String(),
	}

	req, err := http.NewRequest("GET", "https://instagram.com"+c.Request.URL.String(), nil)
	if err != nil {
		slog.Error("Failed to prepare request to follow redirects", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusInternalServerError, "server_error.html", nil)
		return
	}

	res, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to follow redirects", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusInternalServerError, "server_error.html", nil)
		return
	}
	res.Body.Close()

	urlSplit := strings.Split(res.Request.URL.String(), "/")
	postId := urlSplit[len(urlSplit)-2]

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
		c.HTML(http.StatusInternalServerError, "server_error.html", nil)
		return
	}

	if videoUrl == "" {
		slog.Warn("Instagram returned an empty video URL. This most likely means the video is age restricted")
		sentry.CaptureMessage("Instagram returned an empty video URL. This most likely means the video is age restricted")
		c.HTML(http.StatusNoContent, "no_url.html", nil)
		return
	}

	remote, err := url.Parse(videoUrl)
	if err != nil {
		slog.Error("Failed to parse CDN video URL", slog.Any("err", err))
		sentry.CaptureException(err)
		c.HTML(http.StatusInternalServerError, "server_error.html", nil)
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
