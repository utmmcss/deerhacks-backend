package controllers

import (
	//"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/helpers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

func GetApplicaton(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	userDiscordId := user.DiscordId

	var application models.Application
	initializers.DB.First(&application, "discord_id = ?", userDiscordId)

	// If application does not exist, create it and add application to DB
	if application.ID == 0 {

		if user.Status != models.Registering && user.Status != models.Admin {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "User is not allowed to create a new application at this time",
			})
			return
		}

		application = models.Application{
			DiscordId: userDiscordId,
		}

		result := initializers.DB.Create(&application)

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create new application",
			})
			return
		}
	}

	// Convert application to response
	applicationResponse := helpers.ToApplicationResponse(application)

	c.JSON(http.StatusOK, applicationResponse)

}
