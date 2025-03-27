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
package utils

import (
	"log/slog"
	"reflect"
	"strings"
)

const (
	context    = `"contextJSON":`
	contextEnd = `,\"gql_data`
)

// The only reason this exists is to get rid of the context key that's adding unnecessary
// complexity to the json.
type rawHtmlData struct {
	Context struct {
		Media struct {
			Dimensions struct {
				Height, Width int
			} `json:"dimensions"`
			ThumbnailURL   string `json:"display_url"`
			IsVideo        bool   `json:"is_video"`
			VideoURL       string `json:"video_url"`
			VideoViewCount int    `json:"video_view_count"`
			Shortcode      string `json:"shortcode"`
		} `json:"media"`
		Permalink        string `json:"media_permalink"`
		MusicAttribution struct {
			ArtistName            string `json:"artist_name"`
			SongName              string `json:"song_name"`
			UsesOriginalAudio     bool   `json:"uses_original_audio"`
			ShouldMuteAudio       bool   `json:"should_mute_audio"`
			ShouldMuteAudioReason string `json:"should_mute_audio_reason"`
			AudioID               string `json:"audio_id"`
		} `json:"clips_music_attribution_info"`
		Caption       string `json:"caption"`
		CommentsCount int    `json:"comments_count"`
		LikesCount    int    `json:"likes_count"`
		ProfileURL    string `json:"profile_url"`
		Username      string `json:"username"`
		VideoViews    int    `json:"video_views"`
	} `json:"context"`
}

type HtmlData struct {
	Shortcode    string      `json:"shortcode" gorm:"primaryKey;index"`
	Permalink    string      `json:"permalink"`
	ThumbnailURL string      `json:"thumbnail_url"`
	IsVideo      bool        `json:"is_video"`
	Title        string      `json:"title"`
	Views        int         `json:"views"`
	Likes        int         `json:"likes"`
	Comments     int         `json:"comments"`
	Video        *VideoData  `json:"video,omitempty" gorm:"serializer:json"`
	Author       *AuthorData `json:"author" gorm:"serializer:json"`
	ExpiresAt    int64       `json:"expires_at"`
}

func (h *HtmlData) CheckNilField(key string) (any, bool) {
	v := reflect.ValueOf(h).Elem()

	f := v.FieldByName(key)
	if !f.IsValid() {
		return nil, false
	}

	if f.Kind() == reflect.Ptr && f.IsNil() {
		return nil, false
	}

	return f.Interface(), true
}

type VideoData struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

type AuthorData struct {
	Username   string `json:"username"`
	ProfileURL string `json:"profile_url"`
}

// Extracts data from HTML. S is the current line being scanner with [bufio.Scanner]
func ExtractHtmlData(s string) (*HtmlData, bool) {
	startIdx := strings.Index(s, context)
	if startIdx == -1 {
		return nil, false
	}

	s = s[startIdx+len(context)+1:]

	endIdx := strings.Index(s, contextEnd)
	if endIdx == -1 {
		return nil, false
	}

	s = s[:endIdx]
	s = strings.NewReplacer(
		`\"`, `"`,
		`\\\/`, `/`,
		`\\`, `\`,
	).Replace(s)
	s += "}"

	var d *rawHtmlData

	if err := json.Unmarshal([]byte(s), &d); err != nil {
		slog.Error("Failed to unmarshal json", slog.Any("err", err))
		return nil, false
	}

	if strings.HasSuffix(d.Context.Permalink, "invalid") {
		return nil, false
	}

	c := &HtmlData{
		Shortcode: d.Context.Media.Shortcode,
		Author: &AuthorData{
			Username:   d.Context.Username,
			ProfileURL: d.Context.ProfileURL,
		},
		Permalink:    d.Context.Permalink,
		ThumbnailURL: d.Context.Media.ThumbnailURL,
		IsVideo:      d.Context.Media.IsVideo,
		Title:        d.Context.Caption,
		Views:        d.Context.Media.VideoViewCount,
		Likes:        d.Context.LikesCount,
		Comments:     d.Context.CommentsCount,
		Video: &VideoData{
			URL:    d.Context.Media.VideoURL,
			Height: d.Context.Media.Dimensions.Height,
			Width:  d.Context.Media.Dimensions.Width,
		},
	}

	return c, true
}
