package utils

import (
	"errors"
	"regexp"
	"strconv"
)

var (
	validSourceTypes = map[string]bool{
		"game":    true,
		"server":  true,
		"payment": true,
	}
	validStates = map[string]bool{
		"win":  true,
		"lose": true,
	}
	amountRegex = regexp.MustCompile(`^\d+(\.\d{1,2})?$`)
)

func ValidateSourceType(sourceType string) error {
	if !validSourceTypes[sourceType] {
		return errors.New("invalid Source-Type header: must be 'game', 'server', or 'payment'")
	}
	return nil
}

func ValidateState(state string) error {
	if !validStates[state] {
		return errors.New("invalid state: must be 'win' or 'lose'")
	}
	return nil
}

func ValidateAmount(amountStr string) error {
	if !amountRegex.MatchString(amountStr) {
		return errors.New("invalid amount format: must be a string with up to 2 decimal places")
	}
	return nil
}

func ValidateUserID(userIDStr string) (int64, error) {
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil || userID <= 0 {
		return 0, errors.New("invalid user ID: must be a positive integer")
	}
	return userID, nil
}

func ParseAmount(amountStr string) (float64, error) {
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return 0, errors.New("invalid amount: cannot parse as number")
	}
	if amount < 0 {
		return 0, errors.New("invalid amount: cannot be negative")
	}
	return amount, nil
}

func FormatBalance(balance float64) string {
	return strconv.FormatFloat(balance, 'f', 2, 64)
}

