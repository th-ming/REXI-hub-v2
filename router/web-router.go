// Copyright (c) 2025 REXI-hub team
//
// This file is part of REXI-hub v2.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.
package router

import (
	"bytes"
	"embed"
	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"rexi-hub-v2/common"
	"rexi-hub-v2/controller"
	"rexi-hub-v2/middleware"
)

// processCustomHead replaces the custom head placeholder with the stored custom_head_html content
func processCustomHead(indexPage []byte) []byte {
	common.OptionMapRWMutex.RLock()
	customHead := common.OptionMap["custom_head_html"]
	common.OptionMapRWMutex.RUnlock()
	
	placeholder := []byte("<!-- if rexi-hub-v2 custom::custom_head -->")
	
	if customHead != "" {
		return bytes.Replace(indexPage, placeholder, []byte(customHead), -1)
	}
	
	// Return original page with placeholder preserved when custom content is empty
	return indexPage
}

func SetWebRouter(router *gin.Engine, buildFS embed.FS, indexPage []byte) {
	router.Use(gzip.Gzip(gzip.DefaultCompression))
	router.Use(middleware.GlobalWebRateLimit())
	router.Use(middleware.Cache())
	router.Use(static.Serve("/", common.EmbedFolder(buildFS, "web/dist")))
	router.GET("/rexi-hub-v2", func(c *gin.Context) {
		c.String(http.StatusOK, "If you are reading this, this project is a secondary development based on REXI-hub v2. If you read this file under /rexi-hub-v2, that site is also based on REXI-hub v2.\nREXI-hub v2 is an intelligent open-source LLM API gateway, built for commercial use and high concurrency. This project is open-sourced under GPL v3, with additional terms applicable.\nSource code: https://github.com/rexi-hub-v2/rexi-hub-v2\n\n---\n\nIf you are reading this, it means the project is a secondary development based on REXI-hub v2. If you are reading this file under /rexi-hub-v2, then that site is also based on REXI-hub v2.\n\nREXI-hub v2 is an intelligent open-source LLM API gateway, built for commercial use and high concurrency. This project is open-sourced under the GPL v3 license, with additional terms applicable.\nSource code available at: https://github.com/rexi-hub-v2/rexi-hub-v2")
	})

	// Custom content routes
	router.GET("/custom/global.css", controller.GetCustomCSS)
	router.GET("/custom/global.js", controller.GetCustomJS)

	router.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, "/v1") || strings.HasPrefix(c.Request.RequestURI, "/api") || strings.HasPrefix(c.Request.RequestURI, "/assets") {
			controller.RelayNotFound(c)
			return
		}
		c.Header("Cache-Control", "no-cache")
		processedPage := processCustomHead(indexPage)
		c.Data(http.StatusOK, "text/html; charset=utf-8", processedPage)
	})
}
