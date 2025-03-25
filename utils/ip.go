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
package utils

import (
	"bitwise7/vxinst/flags"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	// Initially has all proxies in order
	queue = flags.Proxies
)

// Returns an IP rotation client. If proxies aren't available returns a normal
// HTTP client with a set timeout
func GetIpRotationClient(timeout int) *http.Client {
	if len((*queue)) <= 1 {
		return &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	}

	// Get the next proxy and move it to the end
	next := (*queue)[0]
	*queue = append((*queue)[1:], next)

	if strings.Contains(next, "localhost") {
		return &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
		}
	}

	proxyUrl, _ := url.Parse(next)

	slog.Debug("Using random IP for request", slog.String("ip", proxyUrl.Host))

	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		},
		Timeout: time.Duration(timeout) * time.Second,
	}
}
