package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/antoniomika/Share/handlers"
	"github.com/antoniomika/Share/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Use(utils.CleanupMiddleware)

	r.GET("/", handlers.Index)
	r.GET("/e", handlers.Edit)
	r.GET("/admin", utils.AuthMiddleware, handlers.Admin)

	r.GET("/s/:id", handlers.LoadData)
	r.GET("/u/:id", handlers.LoadData)
	r.GET("/u/:id/:filename", handlers.LoadData)

	apiGroup := r.Group("/api", utils.AuthMiddleware)
	{
		apiGroup.Any("/shorten", handlers.ShortenAPI)
		apiGroup.Any("/upload", handlers.UploadAPI)
		apiGroup.Any("/upload/:filename", handlers.UploadAPI)
	}

	r.NoRoute(func(c *gin.Context) {
		if c.Request.Method != "PUT" {
			c.Redirect(http.StatusFound, os.Getenv("REDIRECT_MAIN"))
		}
		return
	}, utils.AuthMiddleware, handlers.UploadAPI)

	listenPort := os.Getenv("PORT")

	if listenPort == "" {
		listenPort = "8080"
	}

	log.Fatal(r.Run(fmt.Sprintf(":%s", listenPort)))
}
