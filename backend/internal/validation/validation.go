package validation

import (
	"net/mail"
	"strings"

	"github.com/google/uuid"
)

func Required(value string) bool {
	return strings.TrimSpace(value) != ""
}

func Email(value string) bool {
	_, err := mail.ParseAddress(value)
	return err == nil && strings.Contains(value, "@")
}

func Password(value string) bool {
	return len(value) >= 6 && len(value) <= 128
}

func UUID(value string) bool {
	_, err := uuid.Parse(value)
	return err == nil
}

func MaxLen(value string, max int) bool {
	return len([]rune(value)) <= max
}

func ApplicationStatus(value string) bool {
	return value == "accepted" || value == "rejected"
}

func ListingStatus(value string) bool {
	return value == "" || value == "open" || value == "closed"
}
