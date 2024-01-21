package discord

import (
	"fmt"
	"time"
)

func JoinGuildTask(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			fmt.Println("Join Guild Task running", time.Now())

			for {
				// Call DequeueUsers
				users, err := DequeueUsers("join")
				if err != nil {
					fmt.Printf("JoinGuildTask - Error dequeuing users: %v\n", err)
					break // Exit the inner loop on error
				}

				// Check if the list is empty
				if len(users) == 0 {
					fmt.Println("JoinGuildTask - No more users to process. Waiting for the next interval.")
					break // Exit the inner loop if no users are left to process
				}

				// Process the dequeued users
				for _, user := range users {
					if user != nil {
						fmt.Printf("JoinGuildTask - Processing user with Discord ID: %s\n", user.DiscordId)
						AddToDiscord(user, false)
						// Sleep for 3 seconds before processing the next user
						time.Sleep(3 * time.Second)
					}
				}
			}

			fmt.Println("Join Guild Task completed", time.Now())

		}
	}
}

func UpdateRoleTask(interval time.Duration) {
	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			fmt.Println("Update Role Task running", time.Now())

			for {
				// Call DequeueUsers
				users, err := DequeueUsers("update")
				if err != nil {
					fmt.Printf("UpdateRoleTask - Error dequeuing users: %v\n", err)
					break // Exit the inner loop on error
				}

				// Check if the list is empty
				if len(users) == 0 {
					fmt.Println("UpdateRoleTask - No more users to process. Waiting for the next interval.")
					break // Exit the inner loop if no users are left to process
				}

				// Process the dequeued users
				for _, user := range users {
					if user != nil {
						fmt.Printf("UpdateRoleTask - Processing user with Discord ID: %s\n", user.DiscordId)
						UpdateGuildUserRole(user, false)
						// Sleep for 3 seconds before processing the next user
						time.Sleep(3 * time.Second)
					}
				}
			}

			fmt.Println("Update Role Task completed", time.Now())

		}
	}
}
