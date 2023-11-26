package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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

	c.JSON(http.StatusOK, gin.H{})
}

func AdminQRCheckIn(c *gin.Context) {

	type QRCheckInContext string

	const (
		REGISTRATION    QRCheckInContext = "registration"
		DAY_1_DINNER    QRCheckInContext = "day_1_dinner"
		DAY_2_BREAKFAST QRCheckInContext = "day_2_breakfast"
		DAY_2_LUNCH     QRCheckInContext = "day_2_lunch"
		DAY_2_DINNER    QRCheckInContext = "day_2_dinner"
		DAY_3_BREAKFAST QRCheckInContext = "day_3_breakfast"
	)

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
		var checkIns []interface{}
		if scannedUser.CheckIns == nil {
			checkIns = make([]interface{}, 0)
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
