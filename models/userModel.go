package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	DiscordId      string `gorm:"unique"`
	Avatar         string
	Name           string
	Email          string `gorm:"unique"`
	Status         string `gorm:"default:pending"`
	QRCode         string `gorm:"unique"`
	InternalStatus string
	InternalNotes  string
	CheckIns       json.RawMessage `gorm:"type:jsonb"`
	AuthToken      string
	RefreshToken   string
	TokenExpiry    string
}
