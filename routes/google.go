package routes

import (
	"context"
	"fmt"

	auth "github.com/DanSmirnov48/techno-trades-go-backend/authentication"
	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/schemas"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var (
	googleOauthConfig *oauth2.Config
)

func init() {
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  "http://localhost:8000/api/v1/auth/google/callback",
		ClientID:     config.GetConfig().GoogleClientId,
		ClientSecret: config.GetConfig().GoogleClientSecret,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}
}

func (endpoint Endpoint) GoogleLogin(c *fiber.Ctx) error {
	url := googleOauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Println(url)
	return c.Redirect(url)
}

func (endpoint Endpoint) GoogleCallback(c *fiber.Ctx) error {
	db := endpoint.DB
	code := c.Query("code")

	// Use the new validation function
	user, err := auth.ValidateAndFetchGoogleUser(context.Background(), googleOauthConfig, endpoint.DB, code)
	if err != nil {
		return c.Status(500).JSON(utils.RequestErr(utils.ERR_INVALID_AUTH, err.Error()))
	}

	// Generate tokens
	access := auth.GenerateAccessToken(user.ID)
	refresh := auth.GenerateRefreshToken()

	// Update user tokens
	user.Access = &access
	user.Refresh = &refresh
	db.Save(&user)

	// Set cookies
	auth.SetAuthCookie(c, auth.AccessToken, access)
	auth.SetAuthCookie(c, auth.RefreshToken, refresh)

	response := schemas.LoginResponseSchema{
		ResponseSchema: SuccessResponse("Login successful"),
		Data:           schemas.TokensResponseSchema{User: user, Access: access, Refresh: refresh},
	}
	return c.Status(201).JSON(response)
}
