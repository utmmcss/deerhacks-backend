package models

import (
	"gorm.io/gorm"
)

type JoinGuildQueue struct {
	gorm.Model
	DiscordId string `gorm:"unique"`
}

func (JoinGuildQueue) TableName() string {
	return "join_guild_queue"
}
