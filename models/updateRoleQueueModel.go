package models

import (
	"gorm.io/gorm"
)

type UpdateRoleQueue struct {
	gorm.Model
	DiscordId string `gorm:"unique"`
}

func (UpdateRoleQueue) TableName() string {
	return "update_role_queue"
}
