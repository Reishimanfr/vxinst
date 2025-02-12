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
	"bash06/vxinstagram/flags"
	"fmt"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type MediaCandidate struct {
	Width  int    `json:"width"`
	Height int    `json:"height"`
	URL    string `json:"url"`
}

type ImageVersions struct {
	Candidates []MediaCandidate `json:"candidates"`
}

type VideoVersion struct {
	Bandwidth int    `json:"bandwidth"`
	Height    int    `json:"height"`
	ID        string `json:"id"`
	Type      int    `json:"type"`
	URL       string `json:"url"`
	Width     int    `json:"width"`
}

type Item struct {
	ImageVersions ImageVersions  `json:"image_versions2"`
	VideoVersions []VideoVersion `json:"video_versions"`
	HasAudio      bool           `json:"has_audio"`
}

type IgResponse struct {
	Items []Item `json:"items"`
}

// Makes a request to the API using the provided cookie to fetch post info.
// Should only be used if scraping HTML fails
func FetchPost(postId string) (*IgResponse, error) {
	if *flags.InstagramCookie == "" {
		return nil, fmt.Errorf("no instagram cookie provided")
	}

	if *flags.InstagramXIGAppID == "" {
		return nil, fmt.Errorf("no instagram x-ig-app-id provided")
	}

	if *flags.InstagramBrowserAgent == "" {
		return nil, fmt.Errorf("invalid instagram browser agent provided")
	}

	baseURL := "https://www.instagram.com/p/" + postId + "?__a=1&__d=dis"

	req, err := http.NewRequest("GET", baseURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", *flags.InstagramBrowserAgent)
	req.Header.Set("Cookie", *flags.InstagramCookie)
	req.Header.Set("X-IG-App-ID", *flags.InstagramXIGAppID)

	// Set headers so we look more like a real browser
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("Origin", "https://www.instagram.com")
	req.Header.Set("Referer", "https://www.instagram.com")

	resp, err := GetIpRotationClient(5).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: status %d", resp.StatusCode)
	}

	var igResp IgResponse
	if err := json.NewDecoder(resp.Body).Decode(&igResp); err != nil {
		return nil, err
	}

	return &igResp, nil
}
