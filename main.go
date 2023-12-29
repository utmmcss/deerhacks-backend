package main

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/controllers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/middleware"
)

// This function runs before main
func init() {
	if os.Getenv("APP_ENV") != "production" {
		initializers.LoadEnvVariables()
	}
	initializers.ConnectToDB()
	initializers.SyncDatabase()
}

func main() {
	r := gin.Default()
	appEnv := os.Getenv("APP_ENV")

	config := cors.DefaultConfig()
	config.AllowCredentials = true
	config.AllowOrigins = []string{"https://deerhacks.ca", "https://2024.deerhacks.ca"}
	config.ExposeHeaders = []string{"Set-Cookie"}
	config.AllowHeaders = append(config.AllowHeaders, "Cookie")
	if appEnv == "development" {
		config.AllowOrigins = []string{"http://localhost:3000"}
	}
	r.Use(cors.New(config))

	// Start email cleanup task
	go controllers.CleanupTableTask(24 * time.Hour)

	r.POST("/user-login", controllers.Login)
	r.GET("/user-get", middleware.RequireAuth, controllers.GetUser)
	r.POST("/user-update", middleware.RequireAuth, controllers.UpdateUser)
	r.POST("/user-logout", middleware.RequireAuth, controllers.LogoutUser)
	r.POST("/email-verify", controllers.VerifyEmail)

	r.POST("/qr-check-in", middleware.RequireAuth, controllers.AdminQRCheckIn)
	r.POST("/admin-user-update", middleware.RequireAuth, controllers.UpdateAdmin)

	r.GET("/application-get", middleware.RequireAuth, controllers.GetApplicaton)
	r.POST("/application-update", middleware.RequireAuth, controllers.UpdateApplication)

	r.GET("/resume-get", middleware.RequireAuth, controllers.GetResume)
	r.POST("/resume-update", middleware.RequireAuth, middleware.ResumeUpdateRateLimit, controllers.UpdateResume)

	r.Run()

	// r.ForwardedByClientIP = true
	r.SetTrustedProxies([]string{"127.0.0.1"})
}
