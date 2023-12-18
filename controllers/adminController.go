package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

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

			// --- Information can be added or Removed as needed ---
			userResponse["country"] = userApp.Country
			userResponse["education"] = userApp.Education
			userResponse["school"] = userApp.School
			userResponse["program"] = userApp.Program

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

			// Include other fields as needed
			usersResponse = append(usersResponse, userResponse)
		}
	}

	// Send the response
	c.JSON(http.StatusOK, gin.H{
		"users": usersResponse,
	})
}
