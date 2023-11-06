package main

import (
	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/controllers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/middleware"
)

// This function runs before main
func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDB()
	initializers.SyncDatabase()
}

func main() {
	r := gin.Default()

	r.POST("/user-login", controllers.Login)
	r.GET("/user-get", middleware.RequireAuth, controllers.GetUser)
	r.POST("/user-update", middleware.RequireAuth, controllers.UpdateUser)
	r.POST("/admin-user-update", middleware.RequireAuth, controllers.UpdateAdmin)

	r.Run()

	// r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})
}
