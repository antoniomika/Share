package handlers

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/antoniomika/Share/utils"
	"github.com/gin-gonic/gin"
)

// LoadData is the handler to load data from a request url
func LoadData(c *gin.Context) {
	ctx := context.Background()
	kind := ""
	var ent interface{}

	dsClient, err := datastore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		utils.ReturnErr(c, err, http.StatusInternalServerError)
	}

	if strings.HasPrefix(c.Request.URL.Path, "/s/") {
		kind = "Link"
		ent = new(utils.LinkObject)
	} else if strings.HasPrefix(c.Request.URL.Path, "/u/") {
		kind = "Upload"
		ent = new(utils.UploadObject)
	}

	id := ""

	stringArr := strings.Split(c.Param("id"), ".")

	id = stringArr[0]

	key := datastore.NameKey(kind, id, nil)

	if err := dsClient.Get(ctx, key, ent); err != nil {
		if err == datastore.ErrNoSuchEntity {
			c.Redirect(http.StatusFound, "/")
			return
		}
		utils.ReturnErr(c, err, 0)
		return
	}

	link, ok := ent.(*utils.LinkObject)

	if ok {
		newLink := link
		newLink.Clicks++
		newLink.Clickers = append(newLink.Clickers, strings.Split(c.Request.Header.Get("X-Forwarded-For"), ", ")[0])

		if _, err := dsClient.Put(ctx, key, newLink); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		}

		c.Redirect(http.StatusFound, link.URL)
		return
	}

	upload, _ := ent.(*utils.UploadObject)
	uploadKey := upload.StorageKey

	newUpload := upload
	newUpload.Clicks++
	newUpload.Clickers = append(newUpload.Clickers, strings.Split(c.Request.Header.Get("X-Forwarded-For"), ", ")[0])

	if _, err := dsClient.Put(ctx, key, newUpload); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	bucketHandle := client.Bucket(os.Getenv("BUCKET_NAME"))
	reader, err := bucketHandle.Object(uploadKey).NewReader(ctx)

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	defer reader.Close()

	c.Header("Content-Type", upload.ContentType.MIME.Value)
	io.Copy(c.Writer, reader)

	return
}
