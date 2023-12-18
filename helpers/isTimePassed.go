package helpers

import (
	"time"
)

func HasTimePassed(timeStr string) (bool, error) {
	// Parse time string
	givenTime, err := time.Parse(time.RFC3339, timeStr)

	if err != nil {
		return false, err
	}

	// Get current time
	currentTime := time.Now()

	// Check if current time is after given time
	return currentTime.After(givenTime), nil

}
