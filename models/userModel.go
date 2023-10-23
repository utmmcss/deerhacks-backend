package models

import (
	"encoding/json"

	"gorm.io/gorm"
)

type Status string

const (
	Pending     Status = "pending"     // Pending Email Verification
	Registering Status = "registering" // Email Verified, Registering for DeerHacks
	Applied     Status = "applied"     // Application Submitted
	Selected    Status = "selected"    // Selected to Attend DeerHacks, Pending Confirmation
	Accepted    Status = "accepted"    // Accepted to Attend DeerHacks
	Rejected    Status = "rejected"    // Application Rejected
	Attended    Status = "attended"    // Signed in at DeerHacks

	Admin     Status = "admin"     // DeerHacks Tech Organizers
	Moderator Status = "moderator" // DeerHacks Moderators
	Volunteer Status = "volunteer" // DeerHacks Volunteers
)

type User struct {
	gorm.Model
	DiscordId      string `gorm:"unique"`
	Avatar         string
	Name           string `gorm:"size:128"`
	Email          string `gorm:"unique;size:128"`
	Status         Status `gorm:"default:pending"`
	QRCode         string `gorm:"unique"`
	InternalStatus string
	InternalNotes  string
	CheckIns       json.RawMessage `gorm:"type:jsonb"`
	AuthToken      string
	RefreshToken   string
	TokenExpiry    string
}
