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
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

func ScrapeFromHTML(postId string) (*HtmlData, error) {
	origin := "https://instagram.com/p/" + postId + "/embed/captioned"

	slog.Debug("Preparing request", slog.String("origin", origin))
	req, err := http.NewRequest("GET", origin, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare HTTP request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	var client *http.Client

	if !*flags.ProxyScrapeHTML {
		client = &http.Client{
			Timeout: 5 * time.Second,
		}
	} else {
		client = GetIpRotationClient(5)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %v", err)
	}

	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 16*1024), 1024*1024)

	slog.Debug("Scanning response body for video url")
	for scanner.Scan() {
		line := scanner.Text()
		if data, ok := ExtractHtmlData(line); ok {
			slog.Debug("Data found!")
			return data, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning response: %v", err)
	}

	return nil, nil
}
