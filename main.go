package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/antoniomika/Share/handlers"
	"github.com/antoniomika/Share/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func main() {
	store := cookie.NewStore([]byte(os.Getenv("SESSION_SECRET")))
	store.Options(sessions.Options{
		MaxAge: 1 * 60 * 60,
		Path:   "/",
	})

	r := gin.Default()
	r.AppEngine = true

	r.LoadHTMLGlob("templates/*")
	r.Use(sessions.Sessions("session", store))
	r.Use(utils.CleanupMiddleware)

	r.GET("/", handlers.Index)
	r.GET("/e", handlers.Edit)
	r.GET("/admin", utils.AuthMiddleware, handlers.Admin)
	r.GET("/login", handlers.Login)
	r.GET("/logout", handlers.Logout)

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
