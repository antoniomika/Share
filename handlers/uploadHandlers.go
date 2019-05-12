package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/antoniomika/Share/utils"
	"github.com/gin-gonic/gin"
	"github.com/h2non/filetype"
	"google.golang.org/api/iterator"
)

func uploadGet(ctx context.Context, dsClient *datastore.Client, c *gin.Context) {
	var uploads []*utils.UploadObject
	keys, err := dsClient.GetAll(ctx, datastore.NewQuery("Upload"), &uploads)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	res := make(map[string]interface{})

	res["status"] = true
	res["keys"] = keys
	res["uploads"] = uploads

	utils.ReturnJSON(c, res, 0)

	return
}

func uploadPost(ctx context.Context, dsClient *datastore.Client, c *gin.Context) {
	randLen, err := strconv.Atoi(os.Getenv("SHORT_URL_SIZE"))
	if err != nil {
		randLen = 6
	}

	token := utils.RandStringBytesMaskImprSrc(randLen)

	bucket := os.Getenv("BUCKET_NAME")

	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	bucketHandle := client.Bucket(bucket)

	filename := ""

	var uploadFile io.Reader
	if c.Request.Method == "POST" {
		uploadedFile, err := c.FormFile("uploadfile")
		uploadFile, err = uploadedFile.Open()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		filename = uploadedFile.Filename
	} else {
		uploadFile = c.Request.Body
		filename = c.Param("filename")

		if filename == "" {
			filename = c.Request.URL.Path[1:]
		}
	}

	storedFile := filename

	it := bucketHandle.Objects(ctx, &storage.Query{
		Prefix: filename,
	})

	exists := 0
	for {
		_, err := it.Next()
		if err == iterator.Done {
			break
		} else {
			exists++
		}
	}

	if exists > 0 {
		storedFile += fmt.Sprintf(".%d", exists)
	}

	wrt := bucketHandle.Object(storedFile).NewWriter(ctx)

	_, err = io.Copy(wrt, uploadFile)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	err = wrt.Close()
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	url := c.GetHeader("X-Forwarded-Proto") + "://" + c.Request.Host + "/u/" + token + "/" + filename

	expireClicks := c.Query("clicks")
	if expireClicks == "" {
		expireClicks = "0"
	}

	expireClicksInt, err := strconv.Atoi(expireClicks)
	if err != nil {
		log.Printf("failed to get convert int: %v\n", err)
	}

	expireTime := c.Query("time")

	duration, err := time.ParseDuration(expireTime)
	if err != nil {
		log.Printf("failed to parse duration: %v\n", err)
	}

	var expireTimeTime time.Time
	if duration != 0 {
		expireTimeTime = time.Now().Add(duration)
	}

	uploaded := new(utils.UploadObject)

	uploaded.StorageKey = storedFile
	uploaded.Clicks = 0
	uploaded.Clickers = make([]string, 0)
	uploaded.Token = token
	uploaded.Filename = storedFile
	uploaded.ShortURL = url
	uploaded.CreateTime = time.Now()
	uploaded.ExpireClicks = expireClicksInt
	uploaded.ExpireTime = expireTimeTime

	uploadedType, err := filetype.MatchReader(uploadFile)
	if err != nil {
		log.Printf("failed to get content-type: %v\n", err)
	}

	uploaded.ContentType = uploadedType

	key := datastore.NameKey("Upload", token, nil)

	if _, err := dsClient.Put(ctx, key, uploaded); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if c.Query("s") != "" {
		c.Header("Content-Type", "text/plain")
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte(url))
	} else {
		res := make(map[string]interface{})

		res["status"] = true
		res["token"] = token
		res["url"] = url
		res["upload"] = uploaded
		res["bucket"] = bucket

		utils.ReturnJSON(c, res, 0)
	}

	return
}

func uploadDelete(ctx context.Context, dsClient *datastore.Client, c *gin.Context) {
	upload := new(utils.UploadObject)

	key := datastore.NameKey("Upload", c.QueryArray("token")[0], nil)

	if err := dsClient.Get(ctx, key, upload); err != nil {
		if err == datastore.ErrNoSuchEntity {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		utils.ReturnErr(c, err, 0)
		return
	}

	bucket := os.Getenv("BUCKET_NAME")

	client, err := storage.NewClient(ctx)
	if err != nil {
		utils.ReturnErr(c, err, 0)
		return
	}
	defer client.Close()

	bucketHandle := client.Bucket(bucket)

	if err := bucketHandle.Object(upload.Filename).Delete(ctx); err != nil {
		utils.ReturnErr(c, err, 0)
		return
	}

	if err := dsClient.Delete(ctx, key); err != nil {
		utils.ReturnErr(c, err, 0)
		return
	}

	res := make(map[string]interface{})

	res["status"] = true

	utils.ReturnJSON(c, res, 0)

	return
}
