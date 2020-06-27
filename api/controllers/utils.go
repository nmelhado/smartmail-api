package controllers

import (
	"strings"
)

// Convert I to 1, S to 5, Z to 2, or O to 0)
func sanitizeSmartID(unsanitizedSmartID string) string {
	replacer := strings.NewReplacer("I", "1", "S", "5", "Z", "2", "O", "0")
	return replacer.Replace(unsanitizedSmartID)
}
