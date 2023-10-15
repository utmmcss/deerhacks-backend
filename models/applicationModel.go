package models

import (
	"gorm.io/gorm"
)

type Application struct {
	gorm.Model
	DiscordId             string `gorm:"unique"`
	IsDraft               bool
	FirstName             string
	LastName              string
	Email                 string `gorm:"unique"`
	PhoneNumber           string `gorm:"unique"`
	IsSubscribed          bool
	Age                   int
	Gender                string
	Pronoun               string
	Ethnicity             string
	Country               string
	City                  string
	Province              string
	EmergencyName         string
	EmergencyNumber       string
	EmergencyRelationship string
	ShirtSize             string
	DietRestriction       []string `gorm:"type:jsonb"`
	AdditionalInfo        string
	Education             string
	School                string
	Program               string
	ResumeLink            string
	ResumeFilename        string
	ResumeHash            string `gorm:"unique"`
	Github                string
	Linkedin              string
	ResumeConsent         bool
	HackathonExperience   string
	DeerhacksExperience   string
	TeamPreference        string
	Interests             []string `gorm:"type:jsonb"`
	DeerhacksPitch        string
	SharedProject         string
	FutureTech            string
	DeerhacksReach        string
	mlhCodeAgreement      bool
	mlhSubscribe          bool
	mlhAuthorize          bool
}
