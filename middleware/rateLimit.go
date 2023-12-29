package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/models"
)

func ResumeUpdateRateLimit(c *gin.Context) {
	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	if user.ResumeUpdateCount >= 3 {
		c.AbortWithStatus(http.StatusTooManyRequests)
		fmt.Println("UpdateResume - User allowed only 3 update requests")
		return
	}
	c.Next()
}
