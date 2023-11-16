package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

// TODO test when bot is online
func AddToDiscord(user *models.User) bool {
	baseURL := "https://discord.com/api/v10/guilds/" + os.Getenv("GUILD_ID") + "/members/" + user.DiscordId

	formData := url.Values{}
	formData.Set("access_token", user.AuthToken)

	req, err := http.NewRequest("PUT", baseURL, bytes.NewBufferString(formData.Encode()))
	if err != nil {
		fmt.Errorf("error forming request (usersController.AddToDiscord): %s", err)
		return false
	}

	req.Header.Set("Authorization", os.Getenv("BOT_TOKEN"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Errorf("failed to add user to DeerHacks server: %s", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 201

}

func FetchDiscordDetails(code string) (*models.DiscordDetails, error) {
	urlStr := "https://discord.com/api/v10/oauth2/token"

	// Data payload
	data := url.Values{}
	data.Set("client_id", os.Getenv("CLIENT_ID"))
	data.Set("client_secret", os.Getenv("CLIENT_SECRET"))
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", os.Getenv("REDIRECT_URI"))

	// Send POST request with URL-encoded data
	resp, err := http.PostForm(urlStr, data)
	if err != nil {
		return nil, fmt.Errorf("error sending request to Discord token API: %s", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %s", err)
	}

	// Check for application level error
	var appError models.DiscordError
	json.Unmarshal(body, &appError)
	if appError.Error != "" {
		return nil, fmt.Errorf("API Error: %s - %s", appError.Error, appError.ErrorDescription)
	}

	// Unmarshal the JSON Response
	var details models.DiscordDetails
	if err := json.Unmarshal(body, &details); err != nil {
		return nil, fmt.Errorf("error unmarshalling response: %v", err)
	}

	return &details, nil
}

func FetchUserDetails(auth_token string) (*models.DiscordUser, error) {
	const urlStr = "https://discord.com/api/v10/users/@me"

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	// Set the Authorization header with the provided auth_token
	req.Header.Set("Authorization", "Bearer "+auth_token)

	// Send the request
	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Check if response status is OK (200)
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to fetch user details: %s ", string(body))
	}

	// Convert JSON response to model
	var user models.DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func Login(c *gin.Context) {

	// Get discord token off req body
	var body struct {
		Token string `json:"token"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})

		return
	}

	// Send API Request to discord requesting token details

	details, err := FetchDiscordDetails(body.Token)

	if err != nil {

		fmt.Println(err)

		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Stale token provided",
		})

		return
	}

	// Send API Request to discord requesting user details
	userDetails, err := FetchUserDetails(details.AccessToken)

	if err != nil {

		fmt.Println(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to obtain user data",
		})

		return
	}

	// Ensure user is verified (helps avoid spam accounts)
	if !userDetails.Verified {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Unverified discord account",
		})

		return
	}

	// Look up requested user

	var user models.User
	initializers.DB.First(&user, "discord_id = ?", userDetails.ID)

	expiry := time.Now().Add(time.Duration(details.ExpiresIn) * time.Second)

	// If user does not exist, create them and add them to discord

	if user.ID == 0 {

		user = models.User{
			DiscordId:    userDetails.ID,
			Avatar:       userDetails.Avatar,
			Username:     userDetails.Username,
			Email:        userDetails.Email,
			QRCode:       uuid.New().String(),
			AuthToken:    details.AccessToken,
			RefreshToken: details.RefreshToken,
			TokenExpiry:  expiry.Format(time.RFC3339),
		}

		result := initializers.DB.Create(&user)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create user",
			})

			return
		}

		AddToDiscord(&user)

	} else {
		// Update tokens, expiry date, Avatar, and Username for existing user
		user.AuthToken = details.AccessToken
		user.RefreshToken = details.RefreshToken
		user.TokenExpiry = expiry.Format(time.RFC3339)

		if user.Avatar != userDetails.Avatar {
			user.Avatar = userDetails.Avatar
		}

		if user.Username != userDetails.Username {
			user.Username = userDetails.Username
		}

		initializers.DB.Save(&user)
	}

	// Generate a jwt token

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create token",
		})

		return
	}

	// Send it back
	c.SetSameSite(http.SameSiteLaxMode)

	domain := ""

	if os.Getenv("APP_ENV") != "development" {
		domain = "deerhacks.ca"
	}

	c.SetCookie("Authorization", tokenString, 3600*24*30, "", domain, true, true)
	c.JSON(http.StatusOK, gin.H{})
}
