package authentication

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"github.com/DanSmirnov48/techno-trades-go-backend/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var cfg = config.GetConfig()
var SECRETKEY = []byte(cfg.SecretKey)

type CookieType string

const (
	AccessToken  CookieType = "accessToken"
	RefreshToken CookieType = "refreshToken"
)

type AccessTokenPayload struct {
	UserId uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

type RefreshTokenPayload struct {
	Data string `json:"data"`
	jwt.RegisteredClaims
}

func GenerateAccessToken(userId uuid.UUID) string {
	expirationTime := time.Now().Add(time.Duration(cfg.AccessTokenExpireMinutes) * time.Minute)
	payload := AccessTokenPayload{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	// Create the JWT string
	tokenString, err := token.SignedString(SECRETKEY)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		log.Fatal("Error Generating Access token: ", err)
	}
	return tokenString
}

func DecodeAccessToken(token string, db *gorm.DB) (*models.User, *string) {
	claims := &AccessTokenPayload{}

	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return SECRETKEY, nil
	})
	tokenErr := "Auth Token is Invalid or Expired!"
	if err != nil {
		return nil, &tokenErr
	}
	if !tkn.Valid {
		return nil, &tokenErr
	}

	// Fetch User model object
	userId := claims.UserId

	user := models.User{ID: userId}
	result := db.Where(user).First(&user)
	if result.Error != nil {
		return nil, &tokenErr
	}
	return &user, nil
}

func GenerateRefreshToken() string {
	expirationTime := time.Now().Add(time.Duration(cfg.RefreshTokenExpireMinutes) * time.Minute)
	payload := RefreshTokenPayload{
		Data: utils.GetRandomString(10),
		RegisteredClaims: jwt.RegisteredClaims{
			// In JWT, the expiry time is expressed as unix milliseconds
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	// Declare the token with the algorithm used for signing, and the claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	// Create the JWT string
	tokenString, err := token.SignedString(SECRETKEY)
	if err != nil {
		// If there is an error in creating the JWT return an internal server error
		log.Fatal("Error Generating Refresh token: ", err)
	}
	return tokenString
}

func DecodeRefreshToken(token string) bool {
	claims := &RefreshTokenPayload{}
	tkn, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return SECRETKEY, nil
	})
	if err != nil {
		return false
	}
	if !tkn.Valid {
		log.Println("Invalid Refresh Token")
		return false
	}
	return true
}

func SetAuthCookie(c *fiber.Ctx, cookieType CookieType, token string) {
	var expirationMinutes int

	switch cookieType {
	case AccessToken:
		expirationMinutes = cfg.AccessTokenExpireMinutes
	case RefreshToken:
		expirationMinutes = cfg.RefreshTokenExpireMinutes
	default:
		expirationMinutes = 60 // Default to 60 minutes if cookieType is not recognized
	}

	// Determine if the request is secure (HTTPS)
	isSecure := false
	if proto, ok := c.GetReqHeaders()["X-Forwarded-Proto"]; ok {
		for _, p := range proto {
			if p == "https" {
				isSecure = true
				break
			}
		}
	}

	// Set the token in a cookie
	c.Cookie(&fiber.Cookie{
		Name:     string(cookieType),
		Value:    token,
		Expires:  time.Now().Add(time.Duration(expirationMinutes) * time.Minute),
		HTTPOnly: true,     // Prevent access to the cookie via JavaScript
		Secure:   isSecure, // Only send cookie over HTTPS
		SameSite: "Strict", // CSRF protection
	})
}

func RemoveAuthCookie(c *fiber.Ctx, cookieType CookieType) {
	// Set the cookie with an expiration time in the past to remove it
	c.Cookie(&fiber.Cookie{
		Name:     string(cookieType),
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Set to a time in the past to expire the cookie
		HTTPOnly: true,                       // Match the settings of the original cookie
		Secure:   false,                      // Update this if your original cookie is secure
		SameSite: "Strict",                   // CSRF protection
	})
}

// GoogleUserInfo represents the structure of user data from Google
type GoogleUserInfo struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	GoogleID  string `json:"id"`
	GivenName string `json:"given_name"`
	Family    string `json:"family_name"`
}

func ValidateAndFetchGoogleUser(ctx context.Context, oauthConfig *oauth2.Config, db *gorm.DB, code string) (*models.User, error) {
	// Exchange code for token
	token, err := oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, errors.New("failed to exchange token")
	}

	// Get user info from Google
	client := oauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, errors.New("failed to get user info")
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch user info from Google")
	}

	// Parse Google user info
	var googleUser GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		return nil, errors.New("failed to parse user info")
	}

	// Validate required fields
	if googleUser.Email == "" || googleUser.GoogleID == "" {
		return nil, errors.New("incomplete user information from Google")
	}

	user := models.User{Email: googleUser.Email}
	db.Take(&user, user)

	return &user, nil
}
