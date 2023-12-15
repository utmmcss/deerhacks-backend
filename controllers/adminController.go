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

func AdminQRCheckIn(c *gin.Context) {

	type QRCheckIn struct {
		QRid    string           `json:"qrId"`
		Context QRCheckInContext `json:"context"`
	}

	userObj, _ := c.Get("user")

	user := userObj.(models.User)

	if user.Status != models.Admin && user.Status != models.Moderator && user.Status != models.Volunteer {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Admin, moderator, or volunteer only",
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

	var bodyData QRCheckIn

	if json.Unmarshal(bodyObj, &bodyData) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})
		return
	}

	switch bodyData.Context {
	case REGISTRATION, DAY_1_DINNER, DAY_2_BREAKFAST, DAY_2_LUNCH, DAY_2_DINNER, DAY_3_BREAKFAST:
		// Valid context, proceed
	default:
		// Invalid context, return an error
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid context",
		})
		return
	}

	//Get the user scanning in
	var scannedUser models.User
	initializers.DB.First(&scannedUser, "qr_code = ?", bodyData.QRid)

	if scannedUser.Status == models.Admin {
		// Return success if scanning in admins
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("%s checked in successfully", scannedUser.Username),
		})
		return
	} else if bodyData.Context != REGISTRATION {
		// Scanning in for food contexts
		var checkIns []QRCheckInContext
		if scannedUser.CheckIns == nil {
			checkIns = make([]QRCheckInContext, 0)
		} else {
			err := json.Unmarshal(scannedUser.CheckIns, &checkIns)
			if err != nil {
				fmt.Println("Error unmarshalling CheckIns:", err)
				return
			}
			//Check if user has already scanned in for this meal
			for _, item := range checkIns {
				if item == bodyData.Context {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"message": fmt.Sprintf("%s could not be checked in: User has already scanned in for this meal", scannedUser.Username),
					})
					return
				}
			}
		}
		if (scannedUser.Status == models.Moderator && len(checkIns) < 3) || (scannedUser.Status == models.Volunteer && len(checkIns) < 2) || (scannedUser.Status == models.Attended) {
			checkIns = append(checkIns, bodyData.Context)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("%s could not be checked in: Status is not valid for food context", scannedUser.Username),
			})
			return
		}

		//Marshal checkIns to save to database
		checkInsData, err := json.Marshal(checkIns)
		if err != nil {
			fmt.Println("Error marshalling CheckIns:", err)
			return
		}
		scannedUser.CheckIns = checkInsData
	} else if user.Status == models.Admin || user.Status == models.Moderator {
		// Scanning in for registration
		if scannedUser.Status == models.Accepted {
			scannedUser.Status = models.Attended
		} else if scannedUser.Status == models.Attended || scannedUser.Status == models.Moderator || scannedUser.Status == models.Volunteer {
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": fmt.Sprintf("%s checked in successfully", scannedUser.Username),
			})
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("%s could not be checked in: Status is not valid for checkin", scannedUser.Username),
			})
			return
		}

	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": fmt.Sprintf("%s could not be checked in: Volunteers are not authorized to scan in for registration contexts", scannedUser.Username),
		})
		return
	}

	err = initializers.DB.Save(&scannedUser).Error
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("%s checked in successfully", scannedUser.Username),
	})

}
