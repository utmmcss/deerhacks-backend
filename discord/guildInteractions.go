package discord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/utmmcss/deerhacks-backend/models"
)

type DiscordRole string

const (
	PendingDiscord     DiscordRole = "1087192865186254999" // Pending Email Verification
	RegisteringDiscord DiscordRole = "295417774182891522"  // Email Verified, Registering for DeerHacks
	AppliedDiscord     DiscordRole = "1192983763995602964" // Application Submitted
	SelectedDiscord    DiscordRole = "1192983889807933490" // Selected to Attend DeerHacks, Pending Confirmation
	AcceptedDiscord    DiscordRole = "1192984014571704330" // Accepted to Attend DeerHacks
	AttendedDiscord    DiscordRole = "1192984114987548722" // Signed in at DeerHacks
	VolunteerDiscord   DiscordRole = "1100893133581070476" // Volunteer at DeerHacks
	DefaultDiscord     DiscordRole = "1085682655326130316" // Status unknown
)

type RateLimitResponse struct {
	Message    string  `json:"message"`
	RetryAfter float64 `json:"retry_after"`
	Global     bool    `json:"global"`
}

func StatusToDiscordRole(s models.Status) DiscordRole {
	switch s {
	case models.Pending:
		return PendingDiscord
	case models.Registering:
		return RegisteringDiscord
	case models.Applied:
		return AppliedDiscord
	case models.Selected:
		return SelectedDiscord
	case models.Accepted:
		return AcceptedDiscord
	case models.Attended:
		return AttendedDiscord
	case models.Volunteer:
		return VolunteerDiscord
	default:
		return DefaultDiscord
	}
}

func UpdateGuildUserRole(user *models.User, retry bool) bool {

	type DiscordMember struct {
		Roles []DiscordRole `json:"roles"`
	}

	baseURL := "https://discord.com/api/v10/guilds/" + os.Getenv("GUILD_ID") + "/members/" + user.DiscordId

	memberData := DiscordMember{
		Roles: []DiscordRole{StatusToDiscordRole(user.Status)},
	}

	jsonData, err := json.Marshal(memberData)
	if err != nil {
		fmt.Printf("Failed to marshal JSON data (guildInteractions.UpdateGuildUserRole): %s\n", err)
		return false
	}

	req, err := http.NewRequest("PATCH", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("error forming request (guildInteractions.UpdateGuildUserRole): %s", err)
		return false
	}
	req.Header.Set("Authorization", "Bot "+os.Getenv("BOT_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Failed to update users role on discord server: %s", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		return true
	} else if resp.StatusCode == 429 && retry == false {
		var rateLimit RateLimitResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body: %s", err)
			return false
		}

		err = json.Unmarshal(body, &rateLimit)
		if err != nil {
			fmt.Printf("Failed to unmarshal rate limit response: %s", err)
			return false
		}
		retryAfter := time.Duration(rateLimit.RetryAfter * float64(time.Second))
		fmt.Printf("(guildInteractions.UpdateGuildUserRole) Rate limited. Retrying after %v seconds...\n", retryAfter)
		go time.AfterFunc(retryAfter, func() {
			UpdateGuildUserRole(user, true)
		})
	}

	return false

}

func AddToDiscord(user *models.User, retry bool) bool {

	type DiscordMember struct {
		AccessToken string        `json:"access_token"`
		Roles       []DiscordRole `json:"roles"`
	}

	baseURL := "https://discord.com/api/v10/guilds/" + os.Getenv("GUILD_ID") + "/members/" + user.DiscordId

	memberData := DiscordMember{
		AccessToken: user.AuthToken,
		Roles:       []DiscordRole{StatusToDiscordRole(user.Status)},
	}

	jsonData, err := json.Marshal(memberData)
	if err != nil {
		fmt.Printf("Failed to marshal JSON data (guildInteractions.AddToDiscord): %s\n", err)
		return false
	}

	req, err := http.NewRequest("PUT", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("error forming request (guildInteractions.AddToDiscord): %s", err)
		return false
	}
	req.Header.Set("Authorization", "Bot "+os.Getenv("BOT_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("failed to add user to DeerHacks server: %s", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 201 {
		return true
	} else if resp.StatusCode == 204 {
		EnqueueUser(user, "update")
		return true
	} else if resp.StatusCode == 429 && retry == false {
		var rateLimit RateLimitResponse
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Failed to read response body: %s", err)
			return false
		}

		err = json.Unmarshal(body, &rateLimit)
		if err != nil {
			fmt.Printf("Failed to unmarshal rate limit response: %s", err)
			return false
		}
		retryAfter := time.Duration(rateLimit.RetryAfter * float64(time.Second))
		fmt.Printf("(guildInteractions.UpdateGuildUserRole) Rate limited. Retrying after %v seconds...\n", retryAfter)
		go time.AfterFunc(retryAfter, func() {
			AddToDiscord(user, true)
		})
	}

	return false

}
