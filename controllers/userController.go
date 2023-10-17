package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

func GetUser(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	// Create a map for the response
	responseMap := make(map[string]interface{})

	// Always include these fields
	responseMap["name"] = user.Name
	responseMap["email"] = user.Email
	responseMap["discord_id"] = user.DiscordId
	responseMap["status"] = user.Status
	responseMap["qr_code"] = user.QRCode

	// Conditionally include the Avatar field
	if user.Avatar != "" {
		responseMap["avatar"] = user.Avatar
	}

	c.JSON(http.StatusOK, gin.H{
		"user": responseMap,
	})
}

// UpdateUser updates the user's information if they are in the registering status and have verified their email.
// It receives a gin context and expects a JSON request body with the following optional fields: name and email.
// If the request body is invalid or the user is not in the registering status, it returns an error response.
// If the update is successful, it returns a success response.
func UpdateUser(c *gin.Context) {

	var UpdateUserBody struct {
		Name  string `json:"name,omitempty"`
		Email string `json:"email,omitempty"`
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

	body := UpdateUserBody
	if json.Unmarshal(bodyObj, &body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})
		return
	}

	// Early return if body info is already in user to avoid unnecessary DB call
	if user.Name == body.Name && user.Email == body.Email {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	
	// Update the user object with the new information (if applicable)
	if body.Name != "" {
		user.Name = body.Name // Could be truncated to first n characters
	}
	if body.Email != "" {
		user.Email = body.Email // Do we want to use mail.ParseAddress() to validate email?
	}

	// Save the updated user object to the database
	if initializers.DB.Save(&user).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{})
}
