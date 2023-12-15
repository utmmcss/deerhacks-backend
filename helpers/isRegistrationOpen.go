package helpers

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func IsRegistrationOpen() (bool, error) {
	// If registration is closed, return error
	cutOffDateStr := os.Getenv("REGISTRATION_CUTOFF")
	if cutOffDateStr == "" {
		return false, fmt.Errorf("REGISTRATION_CUTOFF environment variable not set")
	}
	cutOffDate, err := strconv.ParseInt(cutOffDateStr, 10, 64)
	if err != nil {
		return false, err
	}
	if time.Now().Unix() > cutOffDate {
		return false, nil
	}
	return true, nil
}
