package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/discord"
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
	var checkIns map[QRCheckInContext]int
	err := json.Unmarshal(rawMsg, &checkIns)
	if err != nil {
		fmt.Println("Error unmarshalling internal status:", err)
		return false
	}
	for key, val := range checkIns {
		switch key {
		case REGISTRATION, DAY_1_DINNER, DAY_2_BREAKFAST, DAY_2_LUNCH, DAY_2_DINNER, DAY_3_BREAKFAST:
			if val < 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func UpdateAdmin(c *gin.Context) {

	type UpdateBody struct {
		FirstName      *string         `json:"first_name,omitempty"`
		LastName       *string         `json:"last_name,omitempty"`
		Email          *string         `json:"email,omitempty"`
		Status         models.Status   `json:"status,omitempty"`
		InternalStatus *string         `json:"internal_status,omitempty"`
		InternalNotes  *string         `json:"internal_notes,omitempty"`
		CheckIns       json.RawMessage `json:"check_ins,omitempty"`
	}

	type UserBatch struct {
		DiscordID string     `json:"discord_id,omitempty"`
		Fields    UpdateBody `json:"fields,omitempty"`
	}

	type UpdateAdminBody struct {
		Users []UserBatch `json:"users,omitempty"`
	}

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	if user.Status != models.Admin && user.Status != models.Moderator {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Admins or Moderators only",
		})
		return
	}

	var bodyObj UpdateAdminBody

	// Bind JSON to bodyData
	if err := c.Bind(&bodyObj); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})
		return
	}

	var currUser models.User
	for _, u := range bodyObj.Users {
		initializers.DB.First(&currUser, "discord_id = ?", u.DiscordID)
		// If discord_id does not exist, return error
		if currUser.ID == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		bodyData := UpdateBody{
			FirstName:      &currUser.FirstName,
			LastName:       &currUser.LastName,
			Email:          &currUser.Email,
			Status:         currUser.Status,
			InternalNotes:  &currUser.InternalNotes,
			InternalStatus: &currUser.InternalStatus,
			CheckIns:       currUser.CheckIns,
		}

		if user.Status == models.Admin || (currUser.Status != models.Admin && currUser.Status != models.Moderator) {

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
			currUser.FirstName = *bodyData.FirstName
			currUser.LastName = *bodyData.LastName

			if *bodyData.Email != currUser.Email {
				email, err := helpers.GetValidEmail(*bodyData.Email)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid Email Address",
					})
					return
				}
				currUser.Email = email
			}

			//Make sure moderators cannot update status to admin or moderator
			if user.Status == models.Moderator && (bodyData.Status == models.Admin || bodyData.Status == models.Moderator) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Moderators cannot update status to admin or moderator",
				})
				return
			} else {
				currUser.Status = bodyData.Status
				discord.EnqueueUser(&currUser, "update")
			}

			currUser.InternalNotes = *bodyData.InternalNotes
			currUser.InternalStatus = *bodyData.InternalStatus

			if bodyData.CheckIns != nil {
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

		// If status is changed to selected send an rsvp email
		if bodyData.Status == models.Selected {
			SetupOutboundEmail(&currUser, "rsvp")
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

	// Check 'status' query parameter
	statuses := []string{}
	if c.DefaultQuery("statuses", "") != "" {
		statuses = strings.Split(c.DefaultQuery("statuses", ""), ",")
	}

	// Check 'internal_statuses' query parameter
	internal_statuses := []string{}
	if c.DefaultQuery("internal_statuses", "") != "" {
		internal_statuses = strings.Split(c.DefaultQuery("internal_statuses", ""), ",")
	}

	// Get search query parameter
	search := c.DefaultQuery("search", "")

	validStatuses := map[string]bool{
		"pending":     true,
		"registering": true,
		"applied":     true,
		"selected":    true,
		"accepted":    true,
		"rejected":    true,
		"attended":    true,
		"admin":       true,
		"moderator":   true,
		"volunteer":   true,
		"guest":       true,
	}

	//return error if status is not valid
	for _, status := range statuses {
		if _, ok := validStatuses[status]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid status filter provided",
			})
			return
		}
	}

	// Form query for internal_statuses

	var internalStatusConditions []string
	var queryParams []interface{}

	for _, status := range internal_statuses {

		if status == "empty" {
			internalStatusConditions = append(internalStatusConditions, "(internal_status IS NULL OR internal_status = '')")
		} else if _, ok := validStatuses[status]; ok {
			internalStatusConditions = append(internalStatusConditions, "internal_status = ?")
			queryParams = append(queryParams, status)
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid internal status filter provided",
			})
			return
		}
	}

	// Check for the 'full' query parameter
	full := c.DefaultQuery("full", "false")

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize := 25

	offset := (page - 1) * pageSize // Calculate the offset for the query

	// Modify the database query to apply status filter if provided
	query := initializers.DB.Model(&models.User{})
	if len(statuses) > 0 {
		query = query.Where("status IN (?)", statuses)
	}

	// Modify the database query to apply internal status filter if provided
	if len(internalStatusConditions) > 0 {
		query = query.Where(strings.Join(internalStatusConditions, " OR "), queryParams...)
	}

	// Modify the database query to apply the search filter if provided
	if search != "" {
		query = query.Where(
			"users.discord_id ILIKE ? OR users.first_name ILIKE ? OR users.last_name ILIKE ? OR users.username ILIKE ? OR users.email ILIKE ? OR users.internal_notes ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%",
		)
	}

	// Modify the database query to apply pagination
	var totalUsers int64
	query.Count(&totalUsers) // Get the total count of users

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
		// Add LIMIT and OFFSET to the query
		query = query.Table("users").
			Select("users.*, applications.*").
			Joins("left join applications on applications.discord_id = users.discord_id").
			Order("users.id").
			Limit(pageSize).
			Offset(offset)
		query.Scan(&userApplications)

		// Iterate over the results to construct the response
		for _, userApp := range userApplications {
			userResponse := make(map[string]interface{})

			userResponse["discord_id"] = userApp.User.DiscordId
			userResponse["first_name"] = userApp.FirstName
			userResponse["last_name"] = userApp.LastName
			userResponse["username"] = userApp.Username
			userResponse["email"] = userApp.Email
			userResponse["status"] = userApp.Status
			userResponse["internal_status"] = userApp.InternalStatus
			userResponse["internal_notes"] = userApp.InternalNotes
			userResponse["check_ins"] = userApp.CheckIns
			userResponse["qr_code"] = userApp.QRCode

			// Users without applications
			if userApp.Application.Model.ID == 0 {
				usersResponse = append(usersResponse, userResponse)
				continue
			}

			appResponse := helpers.ToApplicationResponse(userApp.Application)

			userResponse["is_draft"] = appResponse.IsDraft
			userResponse["application"] = appResponse.Application

			// Call the helper function to get resume details
			resumeFilename, resumeLink, err := GetResumeDetails(&userApp.User, &userApp.Application)
			if err != nil {
				// Handle the error appropriately
				fmt.Println("Error getting resume details: ", err)
				continue // or you can handle it differently based on your application's needs
			}

			// Append the resume information to the user response
			userResponse["resume_file_name"] = resumeFilename
			userResponse["resume_link"] = resumeLink

			// Add the response for the current user to the usersResponse slice
			usersResponse = append(usersResponse, userResponse)
		}
	} else {
		// If full=false, just return the basic user info without joining with the application table
		var users []models.User

		query.
			Order("id").
			Limit(pageSize).
			Offset(offset).
			Find(&users)
		for _, user := range users {
			userResponse := make(map[string]interface{})
			userResponse["discord_id"] = user.DiscordId
			userResponse["first_name"] = user.FirstName
			userResponse["last_name"] = user.LastName
			userResponse["username"] = user.Username
			userResponse["email"] = user.Email
			userResponse["status"] = user.Status
			userResponse["internal_status"] = user.InternalStatus
			userResponse["internal_notes"] = user.InternalNotes
			userResponse["check_ins"] = user.CheckIns
			userResponse["qr_code"] = user.QRCode

			usersResponse = append(usersResponse, userResponse)
		}
	}

	// Calculate the total number of pages
	totalPages := int(math.Ceil(float64(totalUsers) / float64(pageSize)))

	// Prepare pagination metadata
	pagination := gin.H{
		"current_page": page,
		"total_pages":  totalPages,
		"total_users":  totalUsers,
	}

	// Send the response with pagination metadata
	c.JSON(http.StatusOK, gin.H{
		"users":      usersResponse,
		"pagination": pagination,
	})

}
func AdminQRCheckIn(c *gin.Context) {

	userObj, _ := c.Get("user")

	user := userObj.(models.User)

	if user.Status != models.Admin && user.Status != models.Moderator && user.Status != models.Volunteer {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Admin, moderator, or volunteer only",
		})
		return
	}

	type QRCheckIn struct {
		QRid    string           `json:"qrId"`
		Context QRCheckInContext `json:"context"`
	}
	var bodyData QRCheckIn

	bodyObj, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Request Body",
		})
		return
	}
	defer c.Request.Body.Close()

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
	// If qr_code does not exist, return error
	if scannedUser.ID == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	if scannedUser.Status == models.Admin {
		// Return success if scanning in admins
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("%s checked in successfully", scannedUser.Username),
		})
		return
	} else if bodyData.Context != REGISTRATION {
		// Scanning in for food contexts
		var checkIns map[QRCheckInContext]int
		if scannedUser.CheckIns == nil {
			checkIns = make(map[QRCheckInContext]int)
		} else {
			err := json.Unmarshal(scannedUser.CheckIns, &checkIns)
			if err != nil {
				fmt.Println("Error unmarshalling CheckIns:", err)
				return
			}
		}

		value, exists := checkIns[bodyData.Context]
		if exists && ((scannedUser.Status == models.Moderator && value < 3) || (scannedUser.Status == models.Volunteer && value < 2)) {
			checkIns[bodyData.Context] += 1
		} else if !exists && (scannedUser.Status == models.Moderator || scannedUser.Status == models.Volunteer || scannedUser.Status == models.Attended || scannedUser.Status == models.Guest) {
			checkIns[bodyData.Context] = 1
		} else if exists {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("%s could not be checked in: Reached maximum number of check ins for this meal", scannedUser.Username),
			})
			return
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": fmt.Sprintf("%s could not be checked in: User status is not valid for food context", scannedUser.Username),
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
			discord.EnqueueUser(&scannedUser, "update")
		} else if scannedUser.Status == models.Attended || scannedUser.Status == models.Moderator || scannedUser.Status == models.Volunteer || scannedUser.Status == models.Guest {
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

	// Save scanned user to database
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
