package controllers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	brevo "github.com/getbrevo/brevo-go/lib"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/utmmcss/deerhacks-backend/helpers"
	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

func CleanupTableTask(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			fmt.Println("Cleanup Email Task running", time.Now())

			var entries []models.UserEmailContext
			err := initializers.DB.Find(&entries).Error

			if err != nil {
				fmt.Println("Cleanup Failed - Failed to find entries")
				return
			}

			// Start transaction
			tx := initializers.DB.Begin()

			if tx.Error != nil {
				fmt.Println("Cleanup Failed - Failed to Begin Transaction")
				return
			}

			entryIDs := make([]uint, 0, len(entries))

			for _, entry := range entries {

				has_time_passed, timeerr := helpers.HasTimePassed(entry.TokenExpiry)

				if timeerr != nil {
					fmt.Println("Cleanup Failed - HasTimePassed returned unexpected result")
					return
				}

				if has_time_passed {
					entryIDs = append(entryIDs, entry.ID)
				}

			}

			txerr := tx.Where("id IN (?)", entryIDs).Delete(&models.UserEmailContext{}).Error

			if txerr != nil {
				fmt.Println("Cleanup Failed - Batch Delete failed")
				tx.Rollback()
				return
			}

			commiterr := tx.Commit().Error

			if commiterr != nil {
				fmt.Println("Cleanup Failed - Failed to Commit Batch Deletion")
				return
			}

			fmt.Println("Cleanup Succeeded at", time.Now())

		}
	}
}

func getTemplateData(context string, user *models.User, entry *models.UserEmailContext) (string, string, string, error) {
	if context == "signup" {

		subject := "[Action Required] Verify email to access DeerHacks dashboard"

		first_name := user.FirstName

		if first_name == "" {
			first_name = user.Username
		}

		url := "https://deerhacks.ca/verify?code=" + entry.Token

		buttonHTMLTemplate := `<a href="%s" style="background-color: white; color: #181818; padding: 1rem 2rem; font-weight: 600; text-align: center; text-decoration: none; border-radius: 0.5rem; margin: auto;">Verify Email</a>`

		buttonToURL := fmt.Sprintf(buttonHTMLTemplate, url)

		formattedStringHTML := fmt.Sprintf(`
			<div style="background: #212121; padding: 3rem 1rem 1rem; box-sizing: border-box;">
				<div style="background: #181818; color: white; width: 100%%; max-width: 500px; margin: auto; padding: 1rem; border-radius: 1rem; box-sizing: border-box;">
					<img src="https://raw.githubusercontent.com/utmmcss/deerhacks/main/public/backgrounds/collage_close.jpg" alt="DeerHacks Banner" style="width: 100%%; height: auto;">
					<h1 style="color: white;">Deer %s,</h1>
					<h2 style="color: white;">Thanks for creating an account with us at DeerHacks!</h2>
					<p style="color: white;">Please click the button below or this link directly: <a href="%s" style="color: white;">%s</a> to verify your email. The link will expire within 24 hours of receiving this email.</p>
					<div style="display: grid; padding: 3rem 0; box-sizing: border-box;">%s</div>
					<p style="color: white;">Happy Hacking,<br>The DeerHacks Team 🦌</p>
				</div>
				<div style="color: white; width: 100%%; max-width: 500px; margin: auto; padding-top: 1rem; box-sizing: border-box;">
					<p style="color: white;">✨ by <a href="https://github.com/anthonytedja" style="color: white;">Anthony Tedja</a> & <a href="https://github.com/Multivalence" style="color: white;">Shiva Mulwani</a></p>
				</div>
			</div>`,
			first_name, url, url, buttonToURL)

		formattedStringTEXT := fmt.Sprintf("Deer %s,\n\n"+
			"Thanks for creating an account with us at DeerHacks!\n\n"+
			"Please click the link below to verify your email. The link will expire within 24 hours of receiving this email.\n\n"+
			"%s\n\n"+ // Using the button HTML here
			"Happy Hacking,\n\n"+
			"DeerHacks Team 🦌",
			first_name, url)

		return subject, formattedStringHTML, formattedStringTEXT, nil

	} else if context == "rsvp" {
		return "", "", "", nil
	} else {
		return "", "", "", fmt.Errorf("invalid context given")
	}
}

func SendOutboundEmail(email string, html_content string, text_content string, subject string, full_name string) {
	cfg := brevo.NewConfiguration()
	apiClient := brevo.NewAPIClient(cfg)

	ctx := context.WithValue(context.Background(), brevo.ContextAPIKey, brevo.APIKey{
		Key: os.Getenv("BREVO_API_KEY"),
	})

	email_template := brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Email: "no-reply@deerhacks.ca", // Replace with your sender email
			Name:  "DeerHacks",
		},
		To: []brevo.SendSmtpEmailTo{{
			Email: email,
			Name:  full_name, // Optional, can be empty
		}},
		HtmlContent: html_content,
		TextContent: text_content,
		Subject:     subject,
	}

	resp, httpResp, err := apiClient.TransactionalEmailsApi.SendTransacEmail(ctx, email_template)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `TransactionalEmailsApi.SendTransacEmail`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", httpResp)
	} else {
		fmt.Fprintf(os.Stdout, "Email sent successfully to %s: %v\n", email, resp)
	}

}

func SetupOutboundEmail(user *models.User, context string) {

	// Status change configuration
	var status_change = ""

	if context == "signup" {
		status_change = "registering"
	} else if context == "rsvp" {
		status_change = "accepted"
	}

	// Look up user to see if they have an existing request already (with same context)

	var entry models.UserEmailContext
	initializers.DB.First(&entry, "discord_id = ? AND context = ?", user.DiscordId, context)

	expiry := time.Now().Add(24 * time.Hour)

	// If user does not exist, create an entry for them

	if entry.ID == 0 {

		entry = models.UserEmailContext{
			DiscordId:    user.DiscordId,
			Token:        uuid.New().String(),
			TokenExpiry:  expiry.Format(time.RFC3339),
			Context:      context,
			StatusChange: status_change,
		}

		result := initializers.DB.Create(&entry)

		if result.Error != nil {
			fmt.Println("SetupOutboundEmail - Failed to create new DB Entry")
			return
		}
	} else {
		// Overwrite previous email verification with new one
		entry.Token = uuid.New().String()
		entry.TokenExpiry = expiry.Format(time.RFC3339)
		err := initializers.DB.Save(&entry).Error

		if err != nil {
			fmt.Println("SetupOutboundEmail - Failed to overwrite existing DB Entry")
			return
		}
	}

	subject, formattedStringHTML, formattedStringTEXT, err := getTemplateData(context, user, &entry)

	if err == nil {
		SendOutboundEmail(user.Email, formattedStringHTML, formattedStringTEXT, subject, user.FirstName+" "+user.LastName)
	}

}

func VerifyEmail(c *gin.Context) {

	// Get token off req body
	var body struct {
		Token string `json:"token"`
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": "invalid",
		})

		return
	}

	// Attempt to find token
	// If not discovered return invalid

	var matchingEntry models.UserEmailContext
	initializers.DB.First(&matchingEntry, "token = ?", body.Token)

	if matchingEntry.ID == 0 {
		fmt.Println("VerifyEmail - Could not find token given in body")
		c.JSON(http.StatusOK, gin.H{
			"status":  "invalid",
			"context": "invalid",
		})
		return
	}
	// If the token is expired return invalid
	has_time_passed, err := helpers.HasTimePassed(matchingEntry.TokenExpiry)

	if err != nil {
		fmt.Println("VerifyEmail - Calling HasTimePassed failed:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "invalid",
			"context": "invalid",
		})
		return
	}

	if has_time_passed {
		fmt.Println("VerifyEmail - Token expired")
		c.JSON(http.StatusOK, gin.H{
			"status":  "expired",
			"context": matchingEntry.Context,
		})
		return
	}

	// Update User Status and Email

	var user models.User
	initializers.DB.First(&user, "discord_id = ?", matchingEntry.DiscordId)

	if user.ID == 0 {
		fmt.Println("VerifyEmail - Failed to get user data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "invalid",
			"context": "invalid",
		})
		return
	}

	user.Status = models.Status(matchingEntry.StatusChange)
	err = initializers.DB.Save(&user).Error

	if err != nil {
		fmt.Println("VerifyEmail - Failed to save user data")
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  "invalid",
			"context": "invalid",
		})
		return
	}

	fmt.Println("VerifyEmail - Verification succeded for User", user.ID)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"context": matchingEntry.Context,
	})

	err = initializers.DB.Delete(&matchingEntry).Error

	if err != nil {
		fmt.Printf("VerifyEmail - An error occured when trying to delete an entry:", err)
	}

}
