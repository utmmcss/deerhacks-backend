package models

import (
	"gorm.io/gorm"
)

type UpdateRoleQueue struct {
	gorm.Model
	DiscordId string
	Status    Status
}
