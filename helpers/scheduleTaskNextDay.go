package helpers

import (
	"fmt"
	"time"
)

func ScheduleTaskNextDay(task func()) {
	// Get the current time
	now := time.Now()

	// Calculate the start of the next day
	nextDay := now.AddDate(0, 0, 1).Truncate(24 * time.Hour)

	// Calculate the duration until the start of the next day
	durationUntilNextDay := nextDay.Sub(now)

	// Schedule the task
	time.AfterFunc(durationUntilNextDay, task)

	fmt.Println("Task scheduled for:", nextDay)
}
