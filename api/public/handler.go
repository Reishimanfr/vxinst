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
package public

import (
	"bitwise7/vxinst/api/internal"
	"bitwise7/vxinst/flags"
	"bitwise7/vxinst/middleware"
	"net/http"
	"time"

	cache "github.com/chenyahui/gin-cache"
	"github.com/chenyahui/gin-cache/persist"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type Handler struct {
	Db     *gorm.DB
	Router *gin.Engine
}

// Attaches middleware and sets endpoint funcs
func NewHandler(db *gorm.DB) *Handler {
	r := gin.New()

	r.Use(
		gin.Recovery(),
		gin.ErrorLogger(),
		middleware.RateLimiterMiddleware(middleware.NewRateLimiter(5, 10)),
		middleware.CorsMiddleware(),
		// sentrygin.New(sentrygin.Options{

		// }),
	)

	r.LoadHTMLGlob("templates/*")

	return &Handler{
		Db:     db,
		Router: r,
	}
}

func (h *Handler) Init() {
	var st persist.CacheStore = persist.NewMemoryStore(time.Minute * 1)
	cacheExpire := time.Minute * time.Duration(*flags.CacheLifetime)

	if *flags.RedisEnable {
		rdb := redis.NewClient(&redis.Options{
			Addr:     *flags.RedisAddr,
			Password: *flags.RedisPasswd,
			DB:       *flags.RedisDB,
		})

		st = persist.NewRedisStore(rdb)
	}

	// Cache is only enabled if we're not in debug mode
	if !*flags.GinLogs {
		h.Router.Use(cache.CacheByRequestURI(st, cacheExpire))
	}

	h.Router.GET("/reel/:id", h.ServeVideo)
	h.Router.GET("/reels/:id", h.ServeVideo)
	h.Router.GET("/p/:id", h.ServeVideo)
	h.Router.GET("/favicon.ico", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	h.Router.GET("/", func(ctx *gin.Context) {
		ctx.HTML(http.StatusOK, "main.html", gin.H{
			"demo": "/assets/demo.png",
		})
	})
	h.Router.GET("/share/:id", h.FollowShare)
	h.Router.GET("/api/getPostDetails", func(c *gin.Context) { internal.GetPostDetails(c, h.Db) })
}
