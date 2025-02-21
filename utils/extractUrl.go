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
	"strings"
)

type ExtractedData struct {
	VideoURL     string
	ThumbnailURL string
	IsVideo      string
	Title        string
	Views        string
	Comments     string
	Likes        string
	Username     string
}

var prefixes = map[string]string{
	"VideoURL":     `\"video_url\":`,
	"ThumbnailURL": `\"display_url\":`,
	"IsVideo":      `\"is_video\":`,
	"Title":        `\"title\":`,
	"Views":        `\"video_views\":`,
	"Comments":     `\"commenter_count\":`,
	"Likes":        `\"likes_count\":`,
	"Username":     `\"username\":`,
}

const (
	quote = `\"`
)

// Extracts the video URL from response
func ExtractUrl(s string) (*ExtractedData, bool) {
	data := &ExtractedData{}
	ok := false

	for key, prefix := range prefixes {
		startIdx := strings.Index(s, prefix)
		if startIdx == -1 {
			continue
		}
		start := startIdx + len(prefix) + 1
		end := strings.Index(s[start:], quote)
		if end == -1 {
			continue
		}

		value := s[start : start+end]
		value = UnescapeJSONString(value)
		value = strings.ReplaceAll(value, `\/`, `/`)

		switch key {
		case "ThumbnailURL":
			data.ThumbnailURL = value[1:]
		case "Title":
			data.Title = value[1:]
		case "Views":
			data.Views = value[1:]
		case "Comments":
			data.Comments = value[1:]
		case "Likes":
			data.Likes = value[1:]
		case "Username":
			data.Username = value[1:]
		case "VideoURL":
			data.VideoURL = value[1:]
		case "IsVideo":
			data.IsVideo = value[1:]
		}

		ok = true
	}

	return data, ok
}
