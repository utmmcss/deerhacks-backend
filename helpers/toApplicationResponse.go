package helpers

import (

	"github.com/utmmcss/deerhacks-backend/models"
)

type InnerApplication struct {
	PhoneNumber           string           `json:"phone_number"`
	IsSubscribed          bool             `json:"is_subscribed"`
	Age                   int              `json:"age"`
	Gender                string           `json:"gender"`
	Pronoun               string           `json:"pronoun"`
	Ethnicity             string           `json:"ethnicity"`
	Country               string           `json:"country"`
	City                  string           `json:"city"`
	Province              string           `json:"province"`
	EmergencyName         string           `json:"emergency_name"`
	EmergencyNumber       string           `json:"emergency_number"`
	EmergencyRelationship string           `json:"emergency_relationship"`
	ShirtSize             models.ShirtSize `json:"shirt_size"`
	DietRestriction       []string         `json:"diet_restriction"`
	AdditionalInfo        string           `json:"additional_info"`
	Education             string           `json:"education"`
	School                string           `json:"school"`
	Program               string           `json:"program"`
	Github                string           `json:"github"`
	Linkedin              string           `json:"linkedin"`
	ResumeConsent         bool             `json:"resume_consent"`
	HackathonExperience   string           `json:"hackathon_experience"`
	DeerhacksExperience   bool             `json:"deerhacks_experience"`
	TeamPreference        string           `json:"team_preference"`
	Interests             []string         `json:"interests"`
	DeerhacksPitch        string           `json:"deerhacks_pitch"`
	SharedProject         string           `json:"shared_project"`
	FutureTech            string           `json:"future_tech"`
	DeerhacksReach        string           `json:"deerhacks_reach"`
	MlhCodeAgreement      bool             `json:"mlh_code_agreement"`
	MlhSubscribe          bool             `json:"mlh_subscribe"`
	MlhAuthorize          bool             `json:"mlh_authorize"`
}

type ApplicationResponse struct {
	IsDraft     bool             `json:"is_draft"`
	Application InnerApplication `json:"application"`
}

func ToApplicationResponse(application models.Application) ApplicationResponse {

	return ApplicationResponse{
		IsDraft: application.IsDraft,
		Application: InnerApplication{
			PhoneNumber:           application.PhoneNumber,
			IsSubscribed:          application.IsSubscribed,
			Age:                   application.Age,
			Gender:                application.Gender,
			Pronoun:               application.Pronoun,
			Ethnicity:             application.Ethnicity,
			Country:               application.Country,
			City:                  application.City,
			Province:              application.Province,
			EmergencyName:         application.EmergencyName,
			EmergencyNumber:       application.EmergencyNumber,
			EmergencyRelationship: application.EmergencyRelationship,
			ShirtSize:             application.ShirtSize,
			DietRestriction:       application.DietRestriction,
			AdditionalInfo:        application.AdditionalInfo,
			Education:             application.Education,
			School:                application.School,
			Program:               application.Program,
			Github:                application.Github,
			Linkedin:              application.Linkedin,
			ResumeConsent:         application.ResumeConsent,
			HackathonExperience:   application.HackathonExperience,
			DeerhacksExperience:   application.DeerhacksExperience,
			TeamPreference:        application.TeamPreference,
			Interests:             application.Interests,
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
