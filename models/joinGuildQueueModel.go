package models

import (
	"gorm.io/gorm"
)

type JoinGuildQueue struct {
	gorm.Model
	DiscordId string
	Status    Status
}
