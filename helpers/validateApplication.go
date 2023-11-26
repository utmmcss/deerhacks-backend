package helpers

import (

	"github.com/jackc/pgtype"
	"github.com/utmmcss/deerhacks-backend/models"
)

func isListEmpty(list pgtype.JSONB) bool {

	var strList = []string{}
	list.AssignTo(&strList)

	return len(strList) == 0
}

func ValidateApplication(application models.Application) (bool, string) {
	if application.PhoneNumber == "" || len(application.PhoneNumber) < 10 {
		return false, "Phone number is required"
	}
	if application.Age < 18 {
		return false, "You must be at least 18 years old to apply"
	}
	if application.Gender == "" {
		return false, "Gender is required"
	}
	if application.Pronoun == "" {
		return false, "Pronoun is required"
	}
	if isListEmpty(application.Ethnicity) {
		return false, "Ethnicity is required"
	}
	if application.Country == "" {
		return false, "Country is required"
	}
	if application.City == "" {
		return false, "City is required"
	}
	if application.EmergencyName == "" {
		return false, "Emergency contact name is required"
	}
	if application.EmergencyNumber == "" || len(application.EmergencyNumber) < 10 {
		return false, "Emergency contact number is required"
	}
	if application.EmergencyRelationship == "" {
		return false, "Emergency contact relationship is required"
	}
	if application.ShirtSize == "" {
		return false, "Shirt size is required"
	}
	if isListEmpty(application.DietRestriction) {
		return false, "Diet restrictions field missing"
	}
	if application.Education == "" {
		return false, "Education level is required"
	}
	if application.School == "" {
		return false, "School is required"
	}
	if application.Program == "" {
		return false, "Program is required"
	}
	// TODO: Uncomment when resume upload is implemented
	// if application.ResumeLink == "" {
	// 	return false, "Resume link is required"
	// }
	// if application.ResumeFilename == "" {
	// 	return false, "Resume filename is required"
	// }
	// if application.ResumeHash == nil {
	// 	return false, "Resume hash is required"
	// }
	if !application.ResumeConsent {
		return false, "Resume consent is required"
	}
	if isListEmpty(application.DeerhacksExperience) {
		return false, "Deerhacks experience is required"
	}
	if application.HackathonExperience == "" {
		return false, "Hackathon experience is required"
	}
	if application.TeamPreference == "" {
		return false, "Team preference is required"
	}
	if isListEmpty(application.Interests) {
		return false, "Interests field missing"
	}
	if application.DeerhacksPitch == "" {
		return false, "Deerhacks pitch is required"
	}
	if application.SharedProject == "" {
		return false, "Shared project is required"
	}
	if application.FutureTech == "" {
		return false, "Future tech is required"
	}
	if application.DeerhacksReach == "" {
		return false, "Deerhacks reach is required"
	}
	if !application.MlhCodeAgreement {
		return false, "MLH code agreement is required"
	}
	if !application.MlhAuthorize {
		return false, "MLH authorization is required"
	}
	return true, ""
}
