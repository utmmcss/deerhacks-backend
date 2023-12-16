package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/helpers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

func ResumeGetRateLimit(c *gin.Context) {
	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	// Handles unique cases where the Rate Limit is not defined

	if user.ResumeGetRateLimit == "" {
		user.ResumeGetRateLimit = time.Now().Format(time.RFC3339)
		result := initializers.DB.Save(&user)

		if result.Error != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Next()
		return
	}

	// Check if time has passed
	passed, err := helpers.HasTimePassed(user.ResumeGetRateLimit)

	if err != nil {
		fmt.Println("Error parsing time: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// If the time has passed, allow them through and update the rate limit (+5 seconds)
	// Otherwise, abort

	if passed {
		user.ResumeGetRateLimit = time.Now().Add(6 * time.Second).Format(time.RFC3339)
		result := initializers.DB.Save(&user)

		if result.Error != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Next()
		return
	} else {
		c.AbortWithStatus(http.StatusTooManyRequests)
		return

	}
}

func ResumeUpdateRateLimit(c *gin.Context) {
	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	// Handles unique cases where the Rate Limit is not defined

	if user.ResumeUpdateRateLimit == "" {
		user.ResumeUpdateRateLimit = time.Now().Format(time.RFC3339)
		result := initializers.DB.Save(&user)

		if result.Error != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Next()
		return
	}

	// Check if time has passed
	passed, err := helpers.HasTimePassed(user.ResumeUpdateRateLimit)

	if err != nil {
		fmt.Println("Error parsing time: ", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// If the time has passed, allow them through and update the rate limit (+15 minutes)
	// Otherwise, abort

	if passed {
		user.ResumeUpdateRateLimit = time.Now().Add(15 * time.Minute).Format(time.RFC3339)
		result := initializers.DB.Save(&user)

		if result.Error != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.Next()
		return
	} else {
		c.AbortWithStatus(http.StatusTooManyRequests)
		return

	}
}
