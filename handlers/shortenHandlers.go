package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/antoniomika/Share/utils"
	"github.com/gin-gonic/gin"
)

func shortenGet(ctx context.Context, dsClient *datastore.Client, c *gin.Context) {
	var links []*utils.LinkObject
	keys, err := dsClient.GetAll(ctx, datastore.NewQuery("Link"), &links)
	if err != nil {
		utils.ReturnErr(c, err, 0)
		return
	}

	res := make(map[string]interface{})

	res["status"] = true
	res["keys"] = keys
	res["links"] = links

	utils.ReturnJSON(c, res, 0)

	return
}

func shortenPost(ctx context.Context, dsClient *datastore.Client, c *gin.Context) {
	token := utils.RandStringBytesMaskImprSrc(6)

	url := c.GetHeader("X-Forwarded-Proto") + "://" + c.Request.Host + "/s/" + token

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

	link := new(utils.LinkObject)

	link.Token = token
	link.URL = c.QueryArray("url")[0]
	link.Clicks = 0
	link.Clickers = make([]string, 0)
	link.ShortURL = url
	link.CreateTime = time.Now()
	link.ExpireClicks = expireClicksInt
	link.ExpireTime = expireTimeTime

	key := datastore.NameKey("Link", token, nil)

	if _, err := dsClient.Put(ctx, key, link); err != nil {
		utils.ReturnErr(c, err, 0)
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
		res["link"] = link

		utils.ReturnJSON(c, res, 0)
	}

	return
}

func shortenDelete(ctx context.Context, dsClient *datastore.Client, c *gin.Context) {
	link := new(utils.LinkObject)

	key := datastore.NameKey("Link", c.QueryArray("token")[0], nil)

	if err := dsClient.Get(ctx, key, link); err != nil {
		if err == datastore.ErrNoSuchEntity {
			c.Redirect(http.StatusFound, "/")
			return
		}

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
