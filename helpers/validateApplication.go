package helpers

import (
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/utmmcss/deerhacks-backend/models"
)

func ValidateWordCount(fl validator.FieldLevel) bool {
    value := fl.Field().String()
    words := strings.Fields(value)
	count, err := strconv.Atoi(fl.Param())
	if err != nil {
		return false
	}
    return len(words) <= count
}

func ValidateApplication(application models.Application) (bool, []string) {
	validate := validator.New()
	validate.RegisterValidation("wordcount", ValidateWordCount)
	err := validate.Struct(ToApplicationResponse((application)))

	if err == nil {
		return true, []string{}
	}

	errList := []string{}
	for _, field := range err.(validator.ValidationErrors) {
		switch field.Field() {
		case "PhoneNumber":
			errList = append(errList, "Phone number is invalid")
		case "Age":
			errList = append(errList, "You must be at least 18 years old to apply")
		case "ShirtSize":
			errList = append(errList, "Shirt size is invalid")
		case "EmergencyNumber":
			errList = append(errList, "Emergency contact number is invalid")
		case "DietRestriction":
			errList = append(errList, "Diet restrictions field missing")
		case "ResumeConsent":
			errList = append(errList, "Resume consent must be given")
		case "Interests":
			errList = append(errList, "Interests field missing")
		case "DeerhacksPitch":
			errList = append(errList, "Deerhacks pitch must be 100 words or less")
		case "SharedProject":
			errList = append(errList, "Shared project must be 200 words or less")
		case "FutureTech":
			errList = append(errList, "Future tech must be 200 words or less")
		case "MlhCodeAgreement":
			errList = append(errList, "MLH code of conduct must be agreed to")
		case "MlhAuthorize":
			errList = append(errList, "MLH authorization must be given")
		default:
			errList = append(errList, field.Field()+" is required")
		}
	}
	return false, errList
}
