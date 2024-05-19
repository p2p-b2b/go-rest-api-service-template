package repository

import (
	"regexp"
	"strings"
)

func prettyPrint(query string) string {
	ws := regexp.MustCompile(`\s+`)

	out := strings.ReplaceAll(query, "\n", "")
	out = ws.ReplaceAllString(out, " ")
	out = strings.TrimSpace(out)

	return out
}
