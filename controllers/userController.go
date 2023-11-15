package controllers

import (
	"encoding/json"
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

func GetUserList(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

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
			userResponse["verified"] = userApp.Status

			// Include fields from the Application struct
			userResponse["country"] = userApp.Country
			// Add other application fields as needed

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
			userResponse["verified"] = user.Status

			// Include other fields as needed
			usersResponse = append(usersResponse, userResponse)
		}
	}

	// Send the response
	c.JSON(http.StatusOK, gin.H{
		"users": usersResponse,
	})
}
