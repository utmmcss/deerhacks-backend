package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
