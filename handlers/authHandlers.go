package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/datastore"
	"github.com/antoniomika/Share/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	oauth2Config = &oauth2.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		Scopes:       []string{"email", "profile"},
		Endpoint:     google.Endpoint,
	}
)

// Login handles logging the user in to the service
func Login(c *gin.Context) {
	session := sessions.Default(c)

	code := c.Query("code")

	if code == "" {
		url := *c.Request.URL
		url.Host = c.Request.Host
		url.Scheme = c.Request.Header.Get("X-Forwarded-Proto")

		oauth2Config.RedirectURL = url.String()

		if loggedIn := session.Get("loggedin"); loggedIn != nil {
			if loggedIn.(bool) {
				c.Redirect(http.StatusFound, "/admin")
				return
			}
		}

		state := sha256.Sum256(securecookie.GenerateRandomKey(32))

		session.Set("state", base64.URLEncoding.EncodeToString(state[:]))
		session.Save()

		c.Redirect(http.StatusFound, oauth2Config.AuthCodeURL(session.Get("state").(string)))
		return
	}

	if c.Query("state") == session.Get("state") {
		token, err := oauth2Config.Exchange(context.TODO(), code)
		if err != nil {
			log.Println("ISSUE EXCHANGING CODE:", err)
		}

		ctx := context.Background()

		dsClient, err := datastore.NewClient(ctx, os.Getenv("GOOGLE_CLOUD_PROJECT"))
		if err != nil {
			utils.ReturnErr(c, err, http.StatusInternalServerError)
			return
		}

		userData, err := getUserData(*oauth2Config, token)
		if err != nil {
			log.Println("ISSUE GETTING USER DATA:", err)
		}

		user := new(utils.UserObject)
		dsClient.Get(ctx, datastore.NameKey("User", userData["email"].(string), nil), user)

		if user.Email == "" {
			user.Email = userData["email"].(string)

			authToken := sha256.Sum256(securecookie.GenerateRandomKey(32))
			user.AuthToken = base64.URLEncoding.EncodeToString(authToken[:])

			user.Authorized = false
		}

		dsClient.Put(ctx, datastore.NameKey("User", userData["email"].(string), nil), user)

		session.Set("loggedin", user.Authorized)
		session.Save()

		c.Redirect(http.StatusFound, "/admin")
		return
	}

	c.Redirect(http.StatusFound, "/login")
	return
}

// Logout handles logging the user out of the service
func Logout(c *gin.Context) {
	session := sessions.Default(c)

	session.Clear()
	session.Save()

	c.Redirect(http.StatusFound, "/login")
	return
}
