package controllers

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/utmmcss/deerhacks-backend/helpers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
	"gorm.io/gorm"
)

// rename file to specific name
// ensures files are overwritten in S3
const persistentFileName = "Resume.pdf"

func constructS3Key(discordId string) (string, error) {
	appEnv := os.Getenv("APP_ENV")

	folderName := ""
	if appEnv == "development" {
		folderName = "dev"
	} else if appEnv == "production" {
		folderName = "prod"
	} else {
		return "", fmt.Errorf("constructS3Key - environment not defined for current appEnv")
	}

	filepath := folderName + "/" + discordId + "/" + persistentFileName
	return filepath, nil
}

func getPresignedURL(svc *s3.S3, filepath string, filename string) (string, error) {
	// Get the file and return it
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket:                     aws.String("dhapplications"),
		Key:                        aws.String(filepath),
		ResponseContentDisposition: aws.String(fmt.Sprintf("inline; filename=\"%s\"", filename)),
		ResponseContentType:        aws.String("application/pdf"),
	})

	// URL is valid for 7 hours
	urlStr, err := req.Presign(7 * time.Hour)

	if err != nil {
		return "", fmt.Errorf("getPresignedURL -", err)
	}

	return urlStr, nil

}

func GetResumeDetails(user *models.User, application *models.Application) (string, string, error) {
	// If the application or resume link does not exist return empty response
	if application.ID == 0 || application.ResumeLink == "" {
		return "", "", nil
	}

	// Check expiry of resume link
	passed, err := helpers.HasTimePassed(application.ResumeExpiry)
	if err != nil {
		return "", "", fmt.Errorf("error checking resume expiry: %w", err)
	}
	// If expiry has not arrived yet, return link and filename
	if !passed {
		return application.ResumeFilename, application.ResumeLink, nil
	}
	// If link is expired generate new presigned URL (check app_env)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	})
	if err != nil {
		return "", "", fmt.Errorf("error creating AWS session: %w", err)
	}

	svc := s3.New(sess)

	filePath, err := constructS3Key(user.DiscordId)
	if err != nil {
		return "", "", fmt.Errorf("error constructing S3 key: %w", err)
	}

	presignedURL, err := getPresignedURL(svc, filePath, application.ResumeFilename)
	if err != nil {
		return "", "", fmt.Errorf("error getting presigned URL: %w", err)
	}

	// Set ResumeLink to new url and update Expiry
	application.ResumeLink = presignedURL
	application.ResumeExpiry = time.Now().Add(7 * time.Hour).Format(time.RFC3339)

	result := initializers.DB.Save(application)
	if result.Error != nil {
		return "", "", fmt.Errorf("error saving application to database: %w", result.Error)
	}
	// Return link and filename
	return application.ResumeFilename, presignedURL, nil
}

func GetResume(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	var application models.Application
	initializers.DB.First(&application, "discord_id = ?", user.DiscordId)

	filename, link, err := GetResumeDetails(&user, &application)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("GetResume - ", err)
		return
	}

	if filename == "" || link == "" {
		c.JSON(http.StatusOK, gin.H{})
		fmt.Println("GetResume - Application or Resume Link does not exist")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resume_file_name":    application.ResumeFilename,
		"resume_link":         application.ResumeLink,
		"resume_update_count": user.ResumeUpdateCount,
	})
}

func UpdateResume(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	// If user is not registering, return error
	// Admins can update resumes at any time
	if (user.Status != models.Registering) && user.Status != models.Admin {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "User is not allowed to update resume at this time",
		})
		return
	}

	// Ensure registration open
	isOpen, err := helpers.IsRegistrationOpen()

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Failed to check registration status:", err)
		return
	}

	if !isOpen {
		c.AbortWithStatus(http.StatusForbidden)
		fmt.Println("UpdateResume - Registration is closed")
		return
	}

	// Retrieve the file from the posted form-data
	file, err := c.FormFile("file")

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		fmt.Println("UpdateResume - No file provided")
		return
	}

	filename := file.Filename
	fileSizeMB := file.Size / (1024 * 1024)

	// ensure size is less than 2 MB and limit file length
	if fileSizeMB > 2 || len(filename) > 100 {
		c.AbortWithStatus(413)
		fmt.Println("UpdateResume - file/filename too large")
		return
	}

	// ensure file is a pdf
	if filepath.Ext(filename) != ".pdf" {
		c.AbortWithStatus(415)
		fmt.Println("UpdateResume - File not supported")
		return
	}

	// Force file name to be a certain name
	// Ensures files are overwritten in S3
	filename = "Resume.pdf"

	var application models.Application
	initializers.DB.First(&application, "discord_id = ?", user.DiscordId)

	if application.ID == 0 {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Application does not exist")
		return
	}

	// Open the file
	uploadedFile, err := file.Open()
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Error in opening file:", err)
		return
	}
	defer uploadedFile.Close()

	// compute sha256 hash of file
	hash := sha256.New()
	if _, err := io.Copy(hash, uploadedFile); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Error in computing hash:", err)
		return
	}
	computedHash := hex.EncodeToString(hash.Sum(nil))

	// If ResumeHash exists, check if its the same as the one in the request
	// If it is, Get Resume like normal

	if computedHash == application.ResumeHash {
		GetResume(c)
		return
	}

	// Reset the file pointer to the beginning of the file before reading it again
	_, err = uploadedFile.Seek(0, 0) // Seek to the start of the file
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Error in seeking file: ", err)
		return
	}

	// Create s3 session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	})

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Error in creating AWS Session: ", err)
		return
	}

	svc := s3.New(sess)

	// Read File data
	fileData, err := io.ReadAll(uploadedFile)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Error in reading file: ", err)
		return
	}

	// a (weak) check to see if uploaded file contains JavaScript
	if bytes.Contains(fileData, []byte("/JS")) {
		c.AbortWithStatus(http.StatusBadRequest)
		fmt.Println("UpdateResume - Upload of resume with JavaScript attempted by user with discord_id", user.DiscordId)
		return
	}

	s3key, err := constructS3Key(user.DiscordId)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("GetResume - environment not defined for current appEnv", err)
		return
	}

	// Upload the file
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("dhapplications"),
		Key:    aws.String(s3key),
		Body:   bytes.NewReader(fileData),
	})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Failed to upload file: ", err)
		return
	}

	// Get presigned url
	presigned_url, err := getPresignedURL(svc, s3key, file.Filename)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Error in getting presigned url: ", err)
		return
	}

	application.ResumeLink = presigned_url
	application.ResumeExpiry = time.Now().Add(7 * time.Hour).Format(time.RFC3339)
	application.ResumeHash = computedHash
	application.ResumeFilename = file.Filename
	user.ResumeUpdateCount += 1
	result := initializers.DB.Transaction(func(tx *gorm.DB) error {
		if appErr := tx.Save(&application).Error; appErr != nil {
			return appErr
		}
		if userErr := tx.Save(&user).Error; userErr != nil {
			return userErr
		}
		return nil
	})
	if result != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Error in saving Resume Data to Database: ", result)
		return
	}

	// Return link and filename

	c.JSON(http.StatusOK, gin.H{
		"resume_file_name":    file.Filename,
		"resume_link":         application.ResumeLink,
		"resume_update_count": user.ResumeUpdateCount,
	})

}
