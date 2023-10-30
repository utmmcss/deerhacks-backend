package models

import (
	"gorm.io/gorm"
)

type ShirtSize string

const (
	ExtraSmall ShirtSize = "xs"
	Small      ShirtSize = "s"
	Medium     ShirtSize = "m"
	Large      ShirtSize = "l"
	ExtraLarge ShirtSize = "xl"
)

type Application struct {
	gorm.Model
	DiscordId             string `gorm:"unique;size:128"`
	IsDraft               bool   `gorm:"default:true"`
	PhoneNumber           string `gorm:"size:128"`
	IsSubscribed          bool
	Age                   int
	Gender                string `gorm:"size:128"`
	Pronoun               string `gorm:"size:128"`
	Ethnicity             string `gorm:"size:128"`
	Country               string `gorm:"size:128"`
	City                  string `gorm:"size:128"`
	Province              string `gorm:"size:128"`
	EmergencyName         string `gorm:"size:128"`
	EmergencyNumber       string `gorm:"size:128"`
	EmergencyRelationship string `gorm:"size:128"`
	ShirtSize             ShirtSize
	DietRestriction       []string `gorm:"type:jsonb"`
	AdditionalInfo        string   `gorm:"size:128"`
	Education             string   `gorm:"size:128"`
	School                string   `gorm:"size:128"`
	Program               string   `gorm:"size:128"`
	ResumeLink            string   `gorm:"size:128"`
	ResumeFilename        string   `gorm:"size:128"`
	ResumeHash            *string  `gorm:"size:128"`
	Github                string   `gorm:"size:128"`
	Linkedin              string   `gorm:"size:128"`
	ResumeConsent         bool
	HackathonExperience   string `gorm:"size:1300"`
	DeerhacksExperience   bool
	TeamPreference        string   `gorm:"size:128"`
	Interests             []string `gorm:"type:jsonb"`
	DeerhacksPitch        string   `gorm:"size:1300"`
	SharedProject         string   `gorm:"size:1300"`
	FutureTech            string   `gorm:"size:1300"`
	DeerhacksReach        string   `gorm:"size:1300"`
	MlhCodeAgreement      bool
	MlhSubscribe          bool
	MlhAuthorize          bool
}
