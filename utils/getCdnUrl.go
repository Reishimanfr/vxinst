package utils

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
)

// POSSIBLY NOT NEEDED
// Attempts to get the URL to the reel directly from the CDN
func GetCdnUrl(postId string) (string, error) {
	origin := "https://instagram.com/p/" + postId + "/embed/captioned"

	slog.Debug("Preparing request", slog.String("origin", origin))
	req, err := http.NewRequest("GET", origin, nil)
	if err != nil {
		return "", fmt.Errorf("failed to prepare HTTP request: %v", err)
	}

	// Set the user agent to firefox on pc so we get the correct stuff
	req.Header.Set("User-Agent", "Mozilla/5.0 (platform; rv:gecko-version) Gecko/gecko-trail Firefox/firefox-version")

	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %v", err)
	}

	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	scanner.Buffer(make([]byte, 16*1024), 1024*1024)

	slog.Debug("Scanning response body for video url")
	for scanner.Scan() {
		line := scanner.Text()
		if url, found := ExtractUrl(line, false); found && url != "" {
			return url, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error scanning response: %v", err)
	}

	return "", nil
}
