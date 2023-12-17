package helpers

import (
	"github.com/go-playground/validator/v10"
	"github.com/utmmcss/deerhacks-backend/models"
)

func ValidateApplication(application models.Application) (bool, []string) {
	validate := validator.New()
	err := validate.Struct(ToApplicationResponse((application)))

	if err == nil {
		return true, []string{}
	}

	errList := []string{}
	for _, field := range err.(validator.ValidationErrors) {
		switch field.Field() {
		case "Age":
			errList = append(errList, "Age is missing or under 18/over 100")
		case "ShirtSize":
			errList = append(errList, "ShirtSize is missing or invalid")
		case "ResumeConsent":
			errList = append(errList, "ResumeConsent must be given")
		case "MlhCodeAgreement":
			errList = append(errList, "MlhCodeAgreement must be agreed to")
		case "MlhAuthorize":
			errList = append(errList, "MlhAuthorize must be given")
		default:
			errList = append(errList, field.Field()+" field is missing or too long")
		}
	}
	return false, errList
}
