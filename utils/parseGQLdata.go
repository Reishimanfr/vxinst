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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type IGResponse struct {
	Items []IGItem `json:"items"`
}

type IGItem struct {
	Code              string      `json:"code"`
	TakenAt           int64       `json:"taken_at"`
	User              IGUser      `json:"user"`
	IsPaidPartnership bool        `json:"is_paid_partnership"`
	ProductType       string      `json:"product_type"`
	Caption           *IGCaption  `json:"caption"`
	LikeCount         int         `json:"like_count"`
	CommentCount      int         `json:"comment_count"`
	ViewCount         int         `json:"view_count"`
	PlayCount         int         `json:"play_count"`
	VideoDuration     float64     `json:"video_duration"`
	Location          interface{} `json:"location"`
	OriginalHeight    int         `json:"original_height"`
	OriginalWidth     int         `json:"original_width"`
	ImageVersions2    struct {
		Candidates interface{} `json:"candidates"`
	} `json:"image_versions2"`
	VideoVersions interface{} `json:"video_versions"`
	CarouselMedia []struct {
		ImageVersions2 struct {
			Candidates interface{} `json:"candidates"`
		} `json:"image_versions2"`
		VideoVersions interface{} `json:"video_versions"`
	} `json:"carousel_media"`
}

type IGUser struct {
	Username      string `json:"username"`
	FullName      string `json:"full_name"`
	ProfilePicUrl string `json:"profile_pic_url"`
	IsVerified    bool   `json:"is_verified"`
}

type IGCaption struct {
	Text string `json:"text"`
}

type CarouselMediaItem struct {
	ImageVersions interface{} `json:"image_versions"`
	VideoVersions interface{} `json:"video_versions"`
}

type IGData struct {
	Code              string              `json:"code"`
	CreatedAt         int64               `json:"created_at"`
	Username          string              `json:"username"`
	FullName          string              `json:"full_name"`
	ProfilePicture    string              `json:"profile_picture"`
	IsVerified        bool                `json:"is_verified"`
	IsPaidPartnership bool                `json:"is_paid_partnership"`
	ProductType       string              `json:"product_type"`
	Caption           string              `json:"caption"`
	LikeCount         int                 `json:"like_count"`
	CommentCount      int                 `json:"comment_count"`
	ViewCount         int                 `json:"view_count"`
	VideoDuration     float64             `json:"video_duration"`
	Location          interface{}         `json:"location"`
	Height            int                 `json:"height"`
	Width             int                 `json:"width"`
	ImageVersions     interface{}         `json:"image_versions"`
	VideoVersions     interface{}         `json:"video_versions"`
	CarouselMedia     []CarouselMediaItem `json:"carousel_media,omitempty"`
}

type VideoVersions struct {
	Bandwidth int    `json:"bandwidth"`
	Height    int    `json:"height"`
	ID        string `json:"id"`
	Type      int    `json:"type"`
	URL       string `json:"url"`
	Width     int    `json:"width"`
}

func ParseGQLData(postId string) (*IGData, error) {
	baseURL := "https://www.instagram.com/p/" + postId + "?__a=1&__d=dis"
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	q := parsedURL.Query()
	q.Set("__a", "1")
	q.Set("__d", "dis")
	parsedURL.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", parsedURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", *flags.InstagramBrowserAgent)
	req.Header.Set("Cookie", *flags.InstagramCookie)
	req.Header.Set("X-IG-App-ID", *flags.InstagramXIGAppID)
	req.Header.Set("Sec-Fetch-Site", "same-origin")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch data: status %d", resp.StatusCode)
	}

	var igResp IGResponse
	if err := json.NewDecoder(resp.Body).Decode(&igResp); err != nil {
		return nil, err
	}

	if len(igResp.Items) == 0 {
		return nil, errors.New("no items found in response")
	}
	item := igResp.Items[0]

	var carouselMedia []CarouselMediaItem
	if item.ProductType == "carousel_container" {
		for _, el := range item.CarouselMedia {
			carouselMedia = append(carouselMedia, CarouselMediaItem{
				ImageVersions: el.ImageVersions2.Candidates,
				VideoVersions: el.VideoVersions,
			})
		}
	}

	viewCount := item.ViewCount
	if viewCount == 0 {
		viewCount = item.PlayCount
	}

	captionText := ""
	if item.Caption != nil {
		captionText = item.Caption.Text
	}

	data := &IGData{
		Code:              item.Code,
		CreatedAt:         item.TakenAt,
		Username:          item.User.Username,
		FullName:          item.User.FullName,
		ProfilePicture:    item.User.ProfilePicUrl,
		IsVerified:        item.User.IsVerified,
		IsPaidPartnership: item.IsPaidPartnership,
		ProductType:       item.ProductType,
		Caption:           captionText,
		LikeCount:         item.LikeCount,
		CommentCount:      item.CommentCount,
		ViewCount:         viewCount,
		VideoDuration:     item.VideoDuration,
		Location:          item.Location,
		Height:            item.OriginalHeight,
		Width:             item.OriginalWidth,
		ImageVersions:     item.ImageVersions2.Candidates,
		VideoVersions:     item.VideoVersions.(any),
		CarouselMedia:     carouselMedia,
	}

	return data, nil
}
