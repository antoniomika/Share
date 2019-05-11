package handlers

import (
	"context"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/antoniomika/Share/utils"

	"github.com/gin-gonic/gin"
)

// ShortenAPI is the api handler for shortening requests
func ShortenAPI(c *gin.Context) {
	ctx := context.Background()

	dsClient, err := datastore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		utils.ReturnErr(c, err, http.StatusInternalServerError)
	}

	switch c.Request.Method {
	case "GET":
		shortenGet(ctx, dsClient, c)
		return
	case "POST":
		shortenPost(ctx, dsClient, c)
		return
	case "DELETE":
		shortenDelete(ctx, dsClient, c)
		return
	case "PUT":
	}
}

// UploadAPI is the api handler for uploading requests
func UploadAPI(c *gin.Context) {
	ctx := context.Background()

	dsClient, err := datastore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		utils.ReturnErr(c, err, http.StatusInternalServerError)
	}

	switch c.Request.Method {
	case "GET":
		uploadGet(ctx, dsClient, c)
		return
	case "PUT":
		fallthrough
	case "POST":
		uploadPost(ctx, dsClient, c)
		return
	case "DELETE":
		uploadDelete(ctx, dsClient, c)
		return
	}
}
