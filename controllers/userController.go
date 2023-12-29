package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/helpers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

func LogoutUser(c *gin.Context) {

	domain := "localhost"

	if os.Getenv("APP_ENV") != "development" {
		domain = "deerhacks.ca"
	}

	c.SetCookie("Authorization", "", 3600*24*30, "", domain, os.Getenv("APP_ENV") != "development", true)
	c.JSON(http.StatusOK, gin.H{})
}

func GetUser(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	// Create a map for the response
	responseMap := make(map[string]interface{})

	// Always include these fields
	responseMap["first_name"] = user.FirstName
	responseMap["last_name"] = user.LastName
	responseMap["username"] = user.Username
	responseMap["email"] = user.Email
	responseMap["discord_id"] = user.DiscordId
	responseMap["status"] = user.Status
	responseMap["qr_code"] = user.QRCode
	responseMap["avatar"] = user.Avatar

	c.JSON(http.StatusOK, gin.H{
		"user": responseMap,
	})
}

// UpdateUser updates the user's information if they are in the registering status and have verified their email.
// It receives a gin context and expects a JSON request body with the following optional fields: name and email.
// If the request body is invalid or the user is not in the registering status, it returns an error response.
// If the update is successful, it returns a success response.
func UpdateUser(c *gin.Context) {

	type UpdateUserBody struct {
		FirstName string `json:"first_name,omitempty"`
		LastName  string `json:"last_name,omitempty"`
		Email     string `json:"email,omitempty"`
	}

	userObj, _ := c.Get("user")

	user := userObj.(models.User)

	// If user is not registering or pending, they cannot update their information
	if user.Status != models.Registering && user.Status != models.Pending {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not allowed to update information at this time",
		})
		return
	}

	// Get the request body
	bodyObj, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})
		return
	}
	defer c.Request.Body.Close()

	// Defaults to user values
	bodyData := UpdateUserBody{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
	}
	if json.Unmarshal(bodyObj, &bodyData) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})
		return
	}

	var isUserChanged bool = false
	var isEmailChanged bool = false
	var new_email string = ""
	// Update the user object with the new information (if applicable)
	if bodyData.FirstName != user.FirstName {
		user.FirstName = bodyData.FirstName
		isUserChanged = true
	}

	if bodyData.LastName != user.LastName {
		user.LastName = bodyData.LastName
		isUserChanged = true
	}

	if bodyData.Email != user.Email {
		email, err := helpers.GetValidEmail(bodyData.Email)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid Email Address",
			})
			return
		}

		if user.EmailChangeCount >= 20 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Exceeded daily limit of email changes",
			})
			return
		}
		new_email = email
		user.EmailChangeCount += 1
		isEmailChanged = true

		if user.Status == models.Registering {
			user.Status = models.Pending
		}
		isUserChanged = true
	}

	if !isUserChanged {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Save the updated user object to the database
	err = initializers.DB.Save(&user).Error
	if err != nil {

		if helpers.IsUniqueViolationError(err) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already in use",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}

	if isEmailChanged {

		if user.EmailChangeCount == 20 {

			go helpers.ScheduleTaskNextDay(func() {
				err := initializers.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("email_change_count", 0).Error

				if err != nil {
					fmt.Printf("Failed to update Email Change Count back to 0")
				}
			})
		}

		go SetupOutboundEmail(&user, new_email, "signup")

	}

	c.JSON(http.StatusOK, gin.H{})
}
