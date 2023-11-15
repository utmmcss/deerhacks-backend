package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/helpers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

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
		user.Email = email
		isUserChanged = true
	}

	if !isUserChanged {
		c.JSON(http.StatusOK, gin.H{})
		return
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
		if (user.Status == models.Moderator && currUser.Status != models.Admin && currUser.Status != models.Moderator) || user.Status == models.Admin {
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
					return
				}
			} else {
				// Handle error from marshaling
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "An Internal Error Occured",
				})
				return
			}
			// Update the user object with the new information (if applicable)
			if bodyData.FirstName != currUser.FirstName {
				currUser.FirstName = bodyData.FirstName
			}
			if bodyData.LastName != currUser.LastName {
				currUser.LastName = bodyData.LastName
			}
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
			if bodyData.Status != currUser.Status {
				currUser.Status = bodyData.Status
			}
			if bodyData.InternalNotes != currUser.InternalNotes {
				currUser.InternalNotes = bodyData.InternalNotes
			}
			if bodyData.InternalStatus != currUser.InternalStatus {
				currUser.InternalStatus = bodyData.InternalStatus
			}
			if !reflect.DeepEqual(bodyData.CheckIns, currUser.CheckIns) {
				currUser.CheckIns = bodyData.CheckIns
			}

			// Save the updated user object to the database
			if initializers.DB.Save(&currUser).Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update user",
				})
				return
			}
		}
		//Clears currUser in preperation for next user info
		currUser = models.User{}
	}

	c.JSON(http.StatusOK, gin.H{})
}
