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
