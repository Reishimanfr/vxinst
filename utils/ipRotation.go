package utils

import (
	"bash06/vxinstagram/flags"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

var (
	// Initially has all proxies in order
	queue = flags.Proxies
)

func GetIpRotationClient() *http.Client {
	// Get the next proxy and move it to the end
	next := (*queue)[0]
	*queue = append((*queue)[1:], next)
	proxyUrl, _ := url.Parse(next)

	slog.Debug("Using random IP for request", slog.String("ip", proxyUrl.Host))

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
		Timeout: 5 * time.Second,
	}
}
