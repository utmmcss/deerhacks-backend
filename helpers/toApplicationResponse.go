package helpers

import (
	"github.com/utmmcss/deerhacks-backend/models"
)

type InnerApplication struct {
	PhoneNumber           string           `json:"phone_number" validate:"required,gte=10"`
	IsSubscribed          bool             `json:"is_subscribed"`
	Age                   int              `json:"age" validate:"required,gte=18"`
	Gender                string           `json:"gender" validate:"required"`
	Pronoun               string           `json:"pronoun" validate:"required"`
	Ethnicity             []string         `json:"ethnicity" validate:"required,gt=0"`
	Country               string           `json:"country" validate:"required"`
	City                  string           `json:"city" validate:"required"`
	Province              string           `json:"province"`
	EmergencyName         string           `json:"emergency_name" validate:"required"`
	EmergencyNumber       string           `json:"emergency_number" validate:"required,gte=10"`
	EmergencyRelationship string           `json:"emergency_relationship" validate:"required"`
	ShirtSize             models.ShirtSize `json:"shirt_size" validate:"required,oneof=xs s m l xl"`
	DietRestriction       []string         `json:"diet_restriction" validate:"required,gt=0"`
	Day1Dinner            bool             `json:"day1_dinner"`
	Day2Breakfast         bool             `json:"day2_breakfast"`
	Day2Lunch             bool             `json:"day2_lunch"`
	Day2Dinner            bool             `json:"day2_dinner"`
	Day3Breakfast         bool             `json:"day3_breakfast"`
	AdditionalInfo        string           `json:"additional_info"`
	Education             string           `json:"education" validate:"required"`
	School                string           `json:"school"  validate:"required"`
	Program               string           `json:"program"  validate:"required"`
	Portfolio             string           `json:"portfolio"`
	Github                string           `json:"github"`
	Linkedin              string           `json:"linkedin"`
	ResumeConsent         bool             `json:"resume_consent" validate:"eq=true"`
	HackathonExperience   string           `json:"hackathon_experience" validate:"required,gt=0"`
	DeerhacksExperience   []string         `json:"deerhacks_experience" validate:"required,gt=0"`
	TeamPreference        string           `json:"team_preference" validate:"required,gt=0"`
	Interests             []string         `json:"interests" validate:"required,gt=0"`
	DeerhacksPitch        string           `json:"deerhacks_pitch" validate:"required,wordcount=100"`
	SharedProject         string           `json:"shared_project" validate:"required,wordcount=200"`
	FutureTech            string           `json:"future_tech" validate:"required,wordcount=200"`
	DeerhacksReach        string           `json:"deerhacks_reach" validate:"required,gt=0"`
	MlhCodeAgreement      bool             `json:"mlh_code_agreement" validate:"eq=true"`
	MlhSubscribe          bool             `json:"mlh_subscribe"`
	MlhAuthorize          bool             `json:"mlh_authorize" validate:"eq=true"`
}

type ApplicationResponse struct {
	IsDraft     bool             `json:"is_draft"`
	Application InnerApplication `json:"application"`
}

func ToApplicationResponse(application models.Application) ApplicationResponse {

	var ethnicity = []string{}
	var dietRestriction = []string{}
	var deerhacksExperience = []string{}
	var interests = []string{}

	application.Ethnicity.AssignTo(&ethnicity)
	application.DietRestriction.AssignTo(&dietRestriction)
	application.DeerhacksExperience.AssignTo(&deerhacksExperience)
	application.Interests.AssignTo(&interests)

	return ApplicationResponse{
		IsDraft: application.IsDraft,
		Application: InnerApplication{
			PhoneNumber:           application.PhoneNumber,
			IsSubscribed:          application.IsSubscribed,
			Age:                   application.Age,
			Gender:                application.Gender,
			Pronoun:               application.Pronoun,
			Ethnicity:             ethnicity,
			Country:               application.Country,
			City:                  application.City,
			Province:              application.Province,
			EmergencyName:         application.EmergencyName,
			EmergencyNumber:       application.EmergencyNumber,
			EmergencyRelationship: application.EmergencyRelationship,
			ShirtSize:             application.ShirtSize,
			DietRestriction:       dietRestriction,
			Day1Dinner:            application.Day1Dinner,
			Day2Breakfast:         application.Day2Breakfast,
			Day2Lunch:             application.Day2Lunch,
			Day2Dinner:            application.Day2Dinner,
			Day3Breakfast:         application.Day3Breakfast,
			AdditionalInfo:        application.AdditionalInfo,
			Education:             application.Education,
			School:                application.School,
			Program:               application.Program,
			Portfolio:             application.Portfolio,
			Github:                application.Github,
			Linkedin:              application.Linkedin,
			ResumeConsent:         application.ResumeConsent,
			HackathonExperience:   application.HackathonExperience,
			DeerhacksExperience:   deerhacksExperience,
			TeamPreference:        application.TeamPreference,
			Interests:             interests,
			DeerhacksPitch:        application.DeerhacksPitch,
			SharedProject:         application.SharedProject,
			FutureTech:            application.FutureTech,
			DeerhacksReach:        application.DeerhacksReach,
			MlhCodeAgreement:      application.MlhCodeAgreement,
			MlhSubscribe:          application.MlhSubscribe,
			MlhAuthorize:          application.MlhAuthorize,
		},
	}
}
