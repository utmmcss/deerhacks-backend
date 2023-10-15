package initializers

import "github.com/utmmcss/deerhacks-backend/models"

func SyncDatabase() {
	user_err := DB.AutoMigrate(&models.User{})
	app_err := DB.AutoMigrate(&models.Application{})

	if user_err != nil || app_err != nil {
		panic("Failed to Synchronize Database")
	}
}
