package utils

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware is the middleware for handling and verifying auth
func AuthMiddleware(c *gin.Context) {
	token := c.GetHeader("X-Authorization")

	if len(token) == 0 {
		token = c.Query("authorization")
	}

	if len(token) > 0 {
		if token != os.Getenv("ADMIN_PASS") {
			res := make(map[string]interface{})
			res["status"] = false

			c.AbortWithStatusJSON(http.StatusUnauthorized, res)
		}

		return
	} else if c.Request.Host == os.Getenv("SECRET_HOSTNAME") {
		c.Request.Host = os.Getenv("SHARE_HOSTNAME")
		return
	} else {
		res := make(map[string]interface{})
		res["status"] = false

		c.AbortWithStatusJSON(http.StatusUnauthorized, res)
		return
	}
}

// CleanupMiddleware is the middleware to clean expired objects
func CleanupMiddleware(c *gin.Context) {
	ctx := context.Background()

	dsClient, err := datastore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		ReturnErr(c, err, http.StatusInternalServerError)
	}

	bucket := os.Getenv("BUCKET_NAME")
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	bucketHandle := client.Bucket(bucket)

	var links []*LinkObject
	linkKeys, err := dsClient.GetAll(ctx, datastore.NewQuery("Link"), &links)
	if err != nil {
		log.Printf("failed to get links %v\n", err)
		return
	}

	for index, link := range links {
		if (!link.ExpireTime.IsZero() && time.Now().Unix() >= link.ExpireTime.Unix()) || (link.ExpireClicks != 0 && link.Clicks >= link.ExpireClicks) {
			if err := dsClient.Delete(ctx, linkKeys[index]); err != nil {
				log.Printf("failed to delete link %v\n", err)
				return
			}
		}
	}

	var uploads []*UploadObject
	uploadKeys, err := dsClient.GetAll(ctx, datastore.NewQuery("Upload"), &uploads)
	if err != nil {
		log.Printf("failed to get uploads %v\n", err)
		return
	}

	for index, upload := range uploads {
		if (!upload.ExpireTime.IsZero() && time.Now().Unix() >= upload.ExpireTime.Unix()) || (upload.ExpireClicks != 0 && upload.Clicks >= upload.ExpireClicks) {
			if err := bucketHandle.Object(upload.Filename).Delete(ctx); err != nil {
				log.Printf("failed to delete upload file %v\n", err)
				return
			}

			if err := dsClient.Delete(ctx, uploadKeys[index]); err != nil {
				log.Printf("failed to delete upload %v\n", err)
				return
			}
		}
	}
}
