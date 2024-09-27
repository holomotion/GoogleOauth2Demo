package main

import (
	"context"
	"errors"
	"google.golang.org/api/option"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/people/v1"
)

var (
	authConfig *oauth2.Config
	token      *oauth2.Token
	client     *http.Client
	service    *people.Service
)

func main() {
	authConfig = NewGoogleOAuthConfig()
	e := gin.New()
	e.GET("callback", OAuth2Callback)
	e.Use(CheckToken)
	e.GET("info", GetName)
	e.Run("localhost:8080")
}

func NewGoogleOAuthConfig() *oauth2.Config {
	config := &oauth2.Config{
		ClientID:     "xxxxxxx-xxxxxx.apps.googleusercontent.com",
		ClientSecret: "xxxxx-xxxxxxxxxx",
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/user.emails.read",
		},
		Endpoint: google.Endpoint,
	}
	return config
}

func CheckToken(ctx *gin.Context) {
	if token == nil {
		ctx.Redirect(http.StatusFound, authConfig.AuthCodeURL("state"))
		ctx.Abort()
	}
}

func GetName(ctx *gin.Context) {
	p, err := service.People.Get("people/me").PersonFields("names,emailAddresses").Do()
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	result := struct {
		Names []*people.Name         `json:"names"`
		Email []*people.EmailAddress `json:"email"`
	}{
		Names: p.Names,
		Email: p.EmailAddresses,
	}
	ctx.JSON(http.StatusOK, result)
}

func OAuth2Callback(ctx *gin.Context) {
	state := ctx.Query("state")
	if state != "state" {
		ctx.AbortWithError(http.StatusUnauthorized, errors.New("invalid csrf token"))
		return
	}
	code := ctx.Query("code")
	var err error
	token, err = authConfig.Exchange(context.Background(), code)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	client = authConfig.Client(context.Background(), token)
	service, _ = people.NewService(ctx, option.WithTokenSource(authConfig.TokenSource(ctx, token)))
	ctx.Redirect(http.StatusFound, "http://localhost:8080/info")
}
