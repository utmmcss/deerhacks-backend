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
)

func getPresignedURL(svc *s3.S3, filepath string) (string, error) {

	// Decide which folder to use based on app environment

	appEnv := os.Getenv("APP_ENV")

	folderName := ""

	if appEnv == "development" {
		folderName = "dev"
	} else if appEnv == "production" {
		folderName = "prod"
	} else {
		return "", fmt.Errorf("getPresignedURL - folder not defined for current appEnv")
	}

	filepath = folderName + "/" + filepath

	// Get the file and return it
	req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String("dhapplications"),
		Key:    aws.String(filepath),
	})

	// URL is valid for 7 hours
	urlStr, err := req.Presign(7 * time.Hour)

	if err != nil {
		return "", fmt.Errorf("getPresignedURL -", err)
	}

	return urlStr, nil

}

func GetResume(c *gin.Context) {

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

	var application models.Application
	initializers.DB.First(&application, "discord_id = ?", user.DiscordId)

	// If the application or resume link does not exist return empty response
	if application.ID == 0 || application.ResumeLink == "" {
		c.AbortWithStatus(http.StatusOK)
		fmt.Println("GetResume - Application or Resume Link does not exist")
		return
	}

	// Check expiry of resume link
	passed, err := helpers.HasTimePassed(application.ResumeExpiry)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("GetResume - Error in checking resume expiry: ", err)
		return
	}

	// If expiry has not arrived yet, return link and filename
	if !passed {
		c.JSON(http.StatusOK, gin.H{
			"resumeFileName": application.ResumeFilename,
			"resumeLink":     application.ResumeLink,
		})
		return
	}

	// If link is expired generate new presigned URL (check app_env)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	})

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("GetResume - Error in creating AWS Session: ", err)
		return
	}

	svc := s3.New(sess)
	filepath := user.DiscordId + "/" + user.FirstName + "_" + "Resume.pdf"

	presigned_url, err := getPresignedURL(svc, filepath)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("GetResume - Error in getting presigned url: ", err)
		return
	}

	// Set ResumeLink to new url and update Expiry
	application.ResumeLink = presigned_url
	application.ResumeExpiry = time.Now().Add(7 * time.Hour).Format(time.RFC3339)
	result := initializers.DB.Save(&application)

	if result.Error != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("GetResume - Error in saving URL to database: ", err)
		return
	}

	// Return link and filename

	c.JSON(http.StatusOK, gin.H{
		"resumeFileName": application.ResumeFilename,
		"resumeLink":     application.ResumeLink,
	})
}

func UpdateResume(c *gin.Context) {

	// Retrieve the file from the posted form-data
	file, err := c.FormFile("file")

	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		fmt.Println("UpdateResume - No file provided")
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

	userObj, _ := c.Get("user")
	user := userObj.(models.User)

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

	// Upload to user bucket using their discord ID
	// Make sure to change folder depending on app_env

	appEnv := os.Getenv("APP_ENV")

	folderName := ""

	if appEnv == "development" {
		folderName = "dev"
	} else if appEnv == "production" {
		folderName = "prod"
	} else {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Could not identify valid appEnv for folder: ", err)
		return
	}

	filepath := folderName + "/" + user.DiscordId + "/" + filename

	// Upload the file
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("dhapplications"),
		Key:    aws.String(filepath),
		Body:   bytes.NewReader(fileData),
	})
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Failed to upload file: ", err)
		return
	}

	// Get presigned url
	filepath = user.DiscordId + "/" + filename
	presigned_url, err := getPresignedURL(svc, filepath)

	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Error in getting presigned url: ", err)
		return
	}

	application.ResumeLink = presigned_url
	application.ResumeExpiry = time.Now().Add(7 * time.Hour).Format(time.RFC3339)
	application.ResumeHash = computedHash
	application.ResumeFilename = file.Filename
	result := initializers.DB.Save(&application)

	if result.Error != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		fmt.Println("UpdateResume - Error in saving Resume Data to Database: ", err)
		return
	}

	// Return link and filename

	c.JSON(http.StatusOK, gin.H{
		"resumeFileName": file.Filename,
		"resumeLink":     application.ResumeLink,
	})

}
