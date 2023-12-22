package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/helpers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

type QRCheckInContext string

const (
	REGISTRATION    QRCheckInContext = "registration"
	DAY_1_DINNER    QRCheckInContext = "day_1_dinner"
	DAY_2_BREAKFAST QRCheckInContext = "day_2_breakfast"
	DAY_2_LUNCH     QRCheckInContext = "day_2_lunch"
	DAY_2_DINNER    QRCheckInContext = "day_2_dinner"
	DAY_3_BREAKFAST QRCheckInContext = "day_3_breakfast"
)

func checkInsValidation(rawMsg json.RawMessage) bool {
	var checkIns []QRCheckInContext
	err := json.Unmarshal(rawMsg, &checkIns)
	if err != nil {
		fmt.Println("Error unmarshalling internal status:", err)
		return false
	}
	for _, item := range checkIns {
		switch item {
		case REGISTRATION, DAY_1_DINNER, DAY_2_BREAKFAST, DAY_2_LUNCH, DAY_2_DINNER, DAY_3_BREAKFAST:
			// Valid check in context
		default:
			return false
		}
	}
	return true
}

func UpdateAdmin(c *gin.Context) {

	type UpdateBody struct {
		FirstName      string          `json:"first_name,omitempty"`
		LastName       string          `json:"last_name,omitempty"`
		Email          string          `json:"email,omitempty"`
		Status         models.Status   `json:"status,omitempty"`
		InternalStatus string          `json:"internal_status,omitempty"`
		InternalNotes  string          `json:"internal_notes,omitempty"`
		CheckIns       json.RawMessage `json:"check_ins,omitempty"`
	}

	type UserBatch struct {
		DiscordID string     `json:"discordID,omitempty"`
		Fields    UpdateBody `json:"fields,omitempty"`
	}

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	if user.Status != models.Admin && user.Status != models.Moderator {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Admins or Moderators only",
		})
		return
	}

	bodyObj, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})
		return
	}
	defer c.Request.Body.Close()

	var userBatch []UserBatch

	if json.Unmarshal(bodyObj, &userBatch) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})
		return
	}

	var currUser models.User
	for _, u := range userBatch {
		initializers.DB.First(&currUser, "discord_id = ?", u.DiscordID)
		// If discord_id does not exist, return error
		if currUser.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}
		if user.Status == models.Admin || (currUser.Status != models.Admin && currUser.Status != models.Moderator) {
			bodyData := UpdateBody{
				FirstName:      currUser.FirstName,
				LastName:       currUser.LastName,
				Email:          currUser.Email,
				Status:         currUser.Status,
				InternalNotes:  currUser.InternalNotes,
				InternalStatus: currUser.InternalStatus,
				CheckIns:       currUser.CheckIns,
			}
			if jsonData, err := json.Marshal(u.Fields); err == nil {
				if err := json.Unmarshal(jsonData, &bodyData); err != nil {
					// Handle error from unmarshaling
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "An Internal Error Occured",
					})
					fmt.Errorf("Failed to unmarshal u.Fields for user %d ", currUser.ID)
					return
				}
			} else {
				// Handle error from marshaling
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "An Internal Error Occured",
				})
				fmt.Errorf("Failed to marshal u.Fields for user %d ", currUser.ID)
				return
			}
			// Update the user object with the new information (if applicable)
			currUser.FirstName = bodyData.FirstName
			currUser.LastName = bodyData.LastName

			if bodyData.Email != currUser.Email {
				email, err := helpers.GetValidEmail(bodyData.Email)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid Email Address",
					})
					return
				}
				currUser.Email = email
			}

			//Make sure moderators cannot update status to admin
			if user.Status == models.Moderator && bodyData.Status == models.Admin {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Moderators cannot update status to admin",
				})
				return
			} else {
				currUser.Status = bodyData.Status
			}

			currUser.InternalNotes = bodyData.InternalNotes
			currUser.InternalStatus = bodyData.InternalStatus

			if !reflect.DeepEqual(bodyData.CheckIns, currUser.CheckIns) {
				if checkInsValidation(bodyData.CheckIns) {
					currUser.CheckIns = bodyData.CheckIns
				} else {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid CheckIns context",
					})
					return
				}

			}

			// Save the updated user object to the database
			if initializers.DB.Save(&currUser).Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update user",
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Moderators cannot update admins or moderators",
			})
			return
		}
		//Clears currUser in preperation for next user info
		currUser = models.User{}
	}

	c.JSON(http.StatusOK, gin.H{})
}

// Get-user-list endpoint code
func GetUserList(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	// if the user is not an admin or moderator, then return nothing
	if user.Status != models.Admin && user.Status != models.Moderator {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Not allowed to view user list",
		})
		return
	}

	// Check for the 'full' query parameter
	full := c.DefaultQuery("full", "false")

	// Create a slice to hold the response for each user
	var usersResponse []map[string]interface{}

	// if full == true in API call, join application and user table and return specific fields
	if full == "true" {
		// Define a struct to hold the joined data from the user and application tables
		type UserApplication struct {
			models.User
			models.Application
		}

		var userApplications []UserApplication
		// Perform a join with the application table and select all fields
		initializers.DB.Table("users").
			Select("users.*, applications.*").
			Joins("left join applications on applications.discord_id = users.discord_id").
			Scan(&userApplications)

		// Iterate over the results to construct the response
		for _, userApp := range userApplications {
			userResponse := make(map[string]interface{})

			// Include fields from the User struct
			// userResponse["id"] = userApp.ID
			userResponse["first_name"] = userApp.FirstName
			userResponse["last_name"] = userApp.LastName
			userResponse["username"] = userApp.Username
			userResponse["email"] = userApp.Email
			userResponse["status"] = userApp.Status
			userResponse["internal_status"] = userApp.InternalStatus
			userResponse["internal_notes"] = userApp.InternalNotes
			userResponse["check_ins"] = userApp.CheckIns
			userResponse["qr_code"] = userApp.QRCode

			// --- Information can be added or Removed as needed ---
			userResponse["is_draft"] = userApp.IsDraft
			userResponse["phone_number"] = userApp.PhoneNumber
			userResponse["is_subscribed"] = userApp.IsSubscribed
			userResponse["age"] = userApp.Age
			userResponse["gender"] = userApp.Gender
			userResponse["pronouns"] = userApp.Pronoun
			// userResponse["ethnicity"] = userApp.Ethnicity
			userResponse["country"] = userApp.Country
			userResponse["city"] = userApp.City
			userResponse["province"] = userApp.Province
			userResponse["emergency_name"] = userApp.EmergencyName
			userResponse["emergency_number"] = userApp.EmergencyNumber
			userResponse["emergency_relationship"] = userApp.EmergencyRelationship
			userResponse["shirt_size"] = userApp.ShirtSize
			// userResponse["diet_restriction"] = userApp.DietRestriction

			// food information
			userResponse["day1_dinner"] = userApp.Day1Dinner
			userResponse["day2_breakfast"] = userApp.Day2Breakfast
			userResponse["day2_lunch"] = userApp.Day2Lunch
			userResponse["day2_dinner"] = userApp.Day2Dinner
			userResponse["day3_breakfast"] = userApp.Day3Breakfast
			userResponse["additional_info"] = userApp.AdditionalInfo

			// school information
			userResponse["education"] = userApp.Education
			userResponse["school"] = userApp.School
			userResponse["program"] = userApp.Program

			// resume and socials
			userResponse["resume_link"] = userApp.ResumeLink
			userResponse["resume_filename"] = userApp.ResumeFilename
			userResponse["portfolio"] = userApp.Portfolio
			userResponse["github"] = userApp.Github
			userResponse["linkedin"] = userApp.Linkedin
			userResponse["resume_consent"] = userApp.ResumeConsent

			// experiences
			userResponse["hackathon_experience"] = userApp.HackathonExperience
			// userResponse["deerhacks_experience"] = userApp.DeerhacksExperience

			// preferences
			userResponse["team_preference"] = userApp.TeamPreference
			// userResponse["interests"] = userApp.Interests

			// answers
			userResponse["deerhacks_pitch"] = userApp.DeerhacksPitch
			userResponse["shared_project"] = userApp.SharedProject
			userResponse["future_tech"] = userApp.FutureTech
			userResponse["deerhacks_reach"] = userApp.DeerhacksReach

			//mlh
			userResponse["mlh_code_agreement"] = userApp.MlhCodeAgreement
			userResponse["mlh_subscribe"] = userApp.MlhSubscribe
			userResponse["mlh_authorize"] = userApp.MlhAuthorize

			usersResponse = append(usersResponse, userResponse)
		}
	} else {
		// If full=false, just return the basic user info without joining with the application table
		var users []models.User
		initializers.DB.Find(&users)
		for _, user := range users {
			userResponse := make(map[string]interface{})
			userResponse["first_name"] = user.FirstName
			userResponse["last_name"] = user.LastName
			userResponse["username"] = user.Username
			userResponse["email"] = user.Email
			userResponse["status"] = user.Status
			userResponse["internal_status"] = user.InternalStatus
			userResponse["internal_notes"] = user.InternalNotes
			userResponse["check_ins"] = user.CheckIns
			userResponse["qr_code"] = user.QRCode

			// Include other fields as needed
			usersResponse = append(usersResponse, userResponse)
		}
	}

	// Send the response
	c.JSON(http.StatusOK, gin.H{
		"users": usersResponse,
	})
}
