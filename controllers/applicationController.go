package controllers

import (
	//"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

func GetApplicaton(c *gin.Context) {

	userObj, _ := c.Get("user")
	userDiscordId := userObj.(models.User).DiscordId

	var application models.Application
	initializers.DB.First(&application, "discord_id = ?", userDiscordId)


	// If application does not exist, create it and add application to DB
	if application.ID == 0 {
		
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


	c.JSON(http.StatusOK, gin.H{
		"is_draft": application.IsDraft,
		"application": gin.H{
			"phone_number": application.PhoneNumber,
			"is_subscribed": application.IsSubscribed,
			"age": application.Age,
			"gender": application.Gender,
			"pronoun": application.Pronoun,
			"ethnicity": application.Ethnicity,
			"country": application.Country,
			"city": application.City,
			"province": application.Province,
			"emergency_name": application.EmergencyName,
			"emergency_number": application.EmergencyNumber,
			"emergency_relationship": application.EmergencyRelationship,
			"shirt_size": application.ShirtSize,
			"diet_restriction": application.DietRestriction,
			"additional_info": application.AdditionalInfo,
			"education": application.Education,
			"school": application.School,
			"program": application.Program,
			"resume_link": application.ResumeLink,
			"resume_filename": application.ResumeFilename, 
			"github": application.Github,
			"linkedin": application.Linkedin,
			"resume_consent": application.ResumeConsent,
			"hackathon_experience": application.HackathonExperience,
			"deerhacks_experience": application.DeerhacksExperience,
			"team_preference": application.TeamPreference,
			"interests": application.Interests,
			"deerhacks_pitch": application.DeerhacksPitch,
			"shared_project": application.SharedProject,
			"future_tech": application.FutureTech,
			"deerhacks_reach": application.DeerhacksReach,
			"mlh_code_agreement": application.MlhCodeAgreement,
			"mlh_subscribe": application.MlhSubscribe,
			"mlh_authorize": application.MlhAuthorize,
		},
	})

}
