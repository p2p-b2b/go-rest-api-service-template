package repository

import "strings"

func Sanity(value string) string {
	// replace ' with '' to escape value string
	return strings.ReplaceAll(value, "'", "''")
}
