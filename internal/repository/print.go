package repository

import (
	"regexp"
	"strings"
)

func prettyPrint(query string) string {
	ws := regexp.MustCompile(`\s+`)

	out := ws.ReplaceAllString(query, " ")
	out = strings.ReplaceAll(out, "\n", "")
	out = strings.TrimSpace(out)

	return out
}
