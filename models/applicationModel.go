package models

import (
	"github.com/jackc/pgtype"
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
	DietRestriction       pgtype.JSONB `gorm:"type:jsonb;default:'[]'"`
	Day1Dinner            bool
	Day2Breakfast         bool
	Day2Lunch             bool
	Day2Dinner            bool
	Day3Breakfast         bool
	AdditionalInfo        string  `gorm:"size:128"`
	Education             string  `gorm:"size:128"`
	School                string  `gorm:"size:128"`
	Program               string  `gorm:"size:128"`
	ResumeLink            string  `gorm:"size:128"`
	ResumeFilename        string  `gorm:"size:128"`
	ResumeHash            *string `gorm:"size:128"`
	Portfolio             string  `gorm:"size:128"`
	Github                string  `gorm:"size:128"`
	Linkedin              string  `gorm:"size:128"`
	ResumeConsent         bool
	HackathonExperience   string       `gorm:"size:128"`
	DeerhacksExperience   string       `gorm:"size:128"`
	TeamPreference        string       `gorm:"size:128"`
	Interests             pgtype.JSONB `gorm:"type:jsonb;default:'[]'"`
	DeerhacksPitch        string       `gorm:"size:1500"`
	SharedProject         string       `gorm:"size:1500"`
	FutureTech            string       `gorm:"size:1500"`
	DeerhacksReach        string       `gorm:"size:1500"`
	MlhCodeAgreement      bool
	MlhSubscribe          bool
	MlhAuthorize          bool
}
