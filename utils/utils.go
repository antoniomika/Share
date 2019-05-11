package utils

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	firebaseConfig = FirebaseConfig{
		APIKey:            os.Getenv("FIREBASE_APIKEY"),
		AuthDomain:        os.Getenv("FIREBASE_AUTHDOMAIN"),
		DatabaseURL:       os.Getenv("FIREBASE_DATABASEURL"),
		ProjectID:         os.Getenv("FIREBASE_PROJECTID"),
		StorageBucket:     os.Getenv("FIREBASE_STORAGEBUCKET"),
		MessagingSenderID: os.Getenv("FIREBASE_MESSAGINGSENDERID"),
		EditorURL:         os.Getenv("EDITOR_HOSTNAME"),
		IPAddress:         "",
	}
)

// GetFirebaseConfig returns the FireBase config with the user's ip address
func GetFirebaseConfig(c *gin.Context) FirebaseConfig {
	config := firebaseConfig
	config.IPAddress = strings.Split(c.Request.Header.Get("X-Forwarded-For"), ", ")[0]

	return config
}

// ReturnErr returns an error when it happens
func ReturnErr(c *gin.Context, err error, code int) {
	if code == 0 {
		code = http.StatusInternalServerError
	}

	c.AbortWithError(http.StatusInternalServerError, err)
	return
}

// ReturnJSON formats a response as JSON
func ReturnJSON(c *gin.Context, data interface{}, status int) {
	if status == 0 {
		status = http.StatusOK
	}

	c.Header("Content-Type", "application/json")
	c.Writer.WriteHeader(status)

	encoder := json.NewEncoder(c.Writer)
	encoder.SetIndent("", "    ")

	encoder.Encode(data)
	return
}

// RandStringBytesMaskImprSrc creates a random string of length n
func RandStringBytesMaskImprSrc(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	var src = rand.NewSource(time.Now().UnixNano())

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
