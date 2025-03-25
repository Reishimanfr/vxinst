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
package internal

import (
	"bitwise7/vxinst/flags"
	"bitwise7/vxinst/utils"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Returns the details of a post
// Example request would be: GET /api/getPostDetails?id=<postId>
func GetPostDetails(c *gin.Context, db *gorm.DB) {
	postId := c.Query("id")

	if postId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No post id provided",
		})
		return
	}

	var data *utils.HtmlData
	create := false

	err := db.
		Model(&utils.HtmlData{}).
		Where("shortcode = ?", postId).
		First(&data).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			slog.Debug("[internal] Post record not found in database. Fetching new data")
		} else {
			slog.Error("[internal] Failed to retrieve post data from database", slog.Any("err", err))
		}

		create = true

		data, err = utils.ScrapeFromHTML(postId)
		fmt.Println(data)
		if err != nil {
			slog.Error("Failed to scrape from HTML", slog.Any("err", err))
			sentry.CaptureException(err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to scrape post data",
			})
			return
		} else if data == nil {
			slog.Debug("No data returned from scraping. Trying to fetch from API")
			igResp, err := utils.FetchPost(postId)
			if err != nil && err.Error()[0:8] != "bad flag" {
				slog.Error("Failed to fetch data from API", slog.Any("err", err))
				sentry.CaptureException(err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to fetch post data",
				})
				return
			} else if igResp != nil && len(igResp.Items) > 0 && len(igResp.Items[0].VideoVersions) > 0 && len(igResp.Items[0].ImageVersions.Candidates) > 0 {
				data = &utils.HtmlData{
					// TODO: fix this not giving enough data
					Video: &utils.VideoData{
						URL: data.Video.URL,
					},
					ThumbnailURL: igResp.Items[0].ImageVersions.Candidates[0].URL,
				}
			}
		}
	}

	if data != nil {
		c.JSON(http.StatusOK, data)
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No data found for post. The post may be private or instagram may be blocking us",
		})
	}

	if create {
		newRecord := &utils.HtmlData{
			Shortcode: postId,
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(24*(*flags.MemoryLifetime))).Unix(),
		}

		if data != nil {
			newRecord = data
			newRecord.ExpiresAt = time.Now().Add(time.Hour * time.Duration(24*(*flags.MemoryLifetime))).Unix()
		}

		if err := db.Model(&utils.HtmlData{}).Create(newRecord).Error; err != nil {
			sentry.CaptureException(err)
			slog.Error("[internal] Failed to save record to memory database", slog.Any("err", err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save post data",
			})
			return
		}
	}
}
