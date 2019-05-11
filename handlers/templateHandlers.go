package handlers

import (
	"net/http"
	"os"

	"github.com/antoniomika/Share/utils"
	"github.com/gin-gonic/gin"
)

// Index is the main index handler
func Index(c *gin.Context) {
	if c.Request.Host == os.Getenv("EDITOR_HOSTNAME") {
		Edit(c)
		return
	}

	c.Redirect(http.StatusFound, os.Getenv("REDIRECT_MAIN"))
}

// Edit is the main edit handler
func Edit(c *gin.Context) {
	firebaseConfig := utils.GetFirebaseConfig(c)
	c.HTML(http.StatusOK, "edit.html", firebaseConfig)
}

// Admin is the main admin handler
func Admin(c *gin.Context) {
	firebaseConfig := utils.GetFirebaseConfig(c)
	c.HTML(http.StatusOK, "admin.html", firebaseConfig)
}
