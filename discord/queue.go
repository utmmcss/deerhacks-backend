package discord

import (
	"fmt"

	"github.com/utmmcss/deerhacks-backend/initializers"
	"github.com/utmmcss/deerhacks-backend/models"
)

func DequeueUsers(event string) ([]*models.User, error) {
	// Dequeues 25 users at a time
	var sqlQuery string
	var discordIDs []string
	var users []*models.User

	switch event {
	case "join":
		sqlQuery = `SELECT discord_id FROM join_guild_queue ORDER BY created_at ASC LIMIT 25 FOR UPDATE SKIP LOCKED`
	case "update":
		sqlQuery = `SELECT discord_id FROM update_role_queue ORDER BY created_at ASC LIMIT 25 FOR UPDATE SKIP LOCKED`
	default:
		return nil, fmt.Errorf("invalid event type")
	}

	// Begin transaction
	tx := initializers.DB.Begin()

	// Select the next 25 items and lock them
	if err := tx.Raw(sqlQuery).Scan(&discordIDs).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// Delete the items
	if event == "join" {
		if err := tx.Where("discord_id IN ?", discordIDs).Unscoped().Delete(&models.JoinGuildQueue{}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else {
		if err := tx.Where("discord_id IN ?", discordIDs).Unscoped().Delete(&models.UpdateRoleQueue{}).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// Find Users
	if len(discordIDs) > 0 {
		if err := initializers.DB.Where("discord_id IN ?", discordIDs).Find(&users).Error; err != nil {
			return nil, err
		}
	}

	return users, nil
}

func EnqueueUser(user *models.User, event string) error {
	if user == nil {
		return fmt.Errorf("nil user provided")
	}

	switch event {
	case "join":
		joinQueueItem := models.JoinGuildQueue{DiscordId: user.DiscordId}
		if err := initializers.DB.Create(&joinQueueItem).Error; err != nil {
			return err
		}
	case "update":
		updateQueueItem := models.UpdateRoleQueue{DiscordId: user.DiscordId}
		if err := initializers.DB.Create(&updateQueueItem).Error; err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid event type")
	}

	return nil
}
