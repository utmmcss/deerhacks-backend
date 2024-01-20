package controllers

import (
	//"encoding/json"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/utmmcss/deerhacks-backend/discord"
	"github.com/utmmcss/deerhacks-backend/helpers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
	"gorm.io/gorm"
)

func GetApplicaton(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	userDiscordId := user.DiscordId

	var application models.Application
	initializers.DB.First(&application, "discord_id = ?", userDiscordId)

	// If application does not exist, create it and add application to DB
	if application.ID == 0 {

		// If registration is closed, return error
		isRegistrationOpen, err := helpers.IsRegistrationOpen()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create application",
			})
			fmt.Fprintln(os.Stderr, err)
			return
		}
		if !isRegistrationOpen {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Applications are now closed",
			})
			return
		}

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
func UpdateApplication(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	userDiscordId := user.DiscordId

	var application models.Application
	initializers.DB.First(&application, "discord_id = ?", userDiscordId)

	// If application does not exist, return error
	if application.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Application not found",
		})
		return
	}

	// If user is not registering or their application is not a draft, return error
	// Admins can update applications at any time
	if (user.Status != models.Registering || !application.IsDraft) && user.Status != models.Admin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "User is not allowed to update application at this time",
		})
		return
	}

	// If registration is closed, return error
	isRegistrationOpen, err := helpers.IsRegistrationOpen()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update application",
		})
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if !isRegistrationOpen {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Applications are now closed",
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

	// Defaults to current application values
	bodyData := helpers.ToApplicationResponse(application)

	if json.Unmarshal(bodyObj, &bodyData) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})
		return
	}

	// If application is unchanged, return
	if reflect.DeepEqual(bodyData, helpers.ToApplicationResponse(application)) {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	// Update the application object with the new information (if applicable)
	// Manually set pgtype.JSONB fields due to type mismatch
	application.IsDraft = bodyData.IsDraft
	application.Ethnicity.Set(bodyData.Application.Ethnicity)
	application.DietRestriction.Set(bodyData.Application.DietRestriction)
	application.DeerhacksExperience.Set(bodyData.Application.DeerhacksExperience)
	application.Interests.Set(bodyData.Application.Interests)

	if copier.Copy(&application, &(bodyData.Application)) != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update application",
		})
		return
	}

	// Validate application if updated as non-draft
	if !application.IsDraft {
		valid, err := helpers.ValidateApplication(application)
		if !valid {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err,
			})
			return
		}

		// Save the updated user and application object to the database
		user.Status = models.Applied
		discord.EnqueueUser(&user, "update")
		initializers.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Save(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update user/application",
				})
				return err
			}
			if err := tx.Save(&application).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update user/application",
				})
				return err
			}
			c.JSON(http.StatusOK, gin.H{})
			return nil
		})
		return
	}

	// Save the updated application object to the database
	if initializers.DB.Save(&application).Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update application",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
