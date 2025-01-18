package repository

import (
	"regexp"
	"strings"
)

// prettyPrint removes comments, newlines, and extra spaces from a query string.
// It is used to make the query string more readable in the logs.
func prettyPrint(query string) string {
	ws := regexp.MustCompile(`\s+`)

	out := regexp.MustCompile(`--.*\n`).ReplaceAllString(query, "")
	out = strings.ReplaceAll(out, "\n", "")
	out = ws.ReplaceAllString(out, " ")
	out = strings.TrimSpace(out)

	return out
}
