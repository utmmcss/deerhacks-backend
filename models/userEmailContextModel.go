package models

import "gorm.io/gorm"

type UserEmailContext struct {
	gorm.Model
	DiscordId    string
	Token        string `gorm:"size:128"`
	Context      string `gorm:"size:20"`
	StatusChange string `gorm:"size:45"`
	TokenExpiry  string
}
