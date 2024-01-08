package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

// EnsureResumeLink checks and updates the resume link if it has expired.
func GetResumeDetails(user *models.User, application *models.Application) (string, string, error) {
	if application.ID == 0 || application.ResumeLink == "" {
		return "", "", nil
	}

	passed, err := HasTimePassed(application.ResumeExpiry)
	if err != nil {
		return "", "", fmt.Errorf("error checking resume expiry: %w", err)
	}

	if !passed {
		return application.ResumeFilename, application.ResumeLink, nil
	}

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return "", "", fmt.Errorf("error creating AWS session: %w", err)
	}

	svc := s3.New(sess)

	filePath, err := ConstructS3Key(user.DiscordId)
	if err != nil {
		return "", "", fmt.Errorf("error constructing S3 key: %w", err)
	}

	presignedURL, err := GetPresignedURL(svc, filePath, application.ResumeFilename)
	if err != nil {
		return "", "", fmt.Errorf("error getting presigned URL: %w", err)
	}

	application.ResumeLink = presignedURL
	application.ResumeExpiry = time.Now().Add(7 * time.Hour).Format(time.RFC3339)

	result := initializers.DB.Save(application)
	if result.Error != nil {
		return "", "", fmt.Errorf("error saving application to database: %w", result.Error)
	}

	return application.ResumeFilename, presignedURL, nil
}

// ConstructS3Key builds the S3 key based on the environment and user's Discord ID.
func ConstructS3Key(discordId string) (string, error) {
	appEnv := os.Getenv("APP_ENV")
	var folderName string
	switch appEnv {
	case "development":
		folderName = "dev"
	case "production":
		folderName = "prod"
	default:
		return "", fmt.Errorf("environment not defined for current appEnv")
	}
	return fmt.Sprintf("%s/%s/%s", folderName, discordId, "Resume.pdf"), nil
}

// GetPresignedURL creates a presigned URL for the given S3 key and filename.
func GetPresignedURL(svc *s3.S3, filePath string, filename string) (string, error) {
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String("dhapplications"),
		Key:                        aws.String(filePath),
		ResponseContentDisposition: aws.String(fmt.Sprintf("attachment; filename=\"%s\"", filename)),
		ResponseContentType:        aws.String("application/pdf"),
	})
	urlStr, err := req.Presign(7 * time.Hour)
	if err != nil {
		return "", fmt.Errorf("error presigning URL: %w", err)
	}
	return urlStr, nil
}
