package repository

import (
	"regexp"
	"strings"
)

func prettyPrint(query string) string {
	ws := regexp.MustCompile(`\s+`)

	out := regexp.MustCompile(`--.*\n`).ReplaceAllString(query, "")
	out = strings.ReplaceAll(out, "\n", "")
	out = ws.ReplaceAllString(out, " ")
	out = strings.TrimSpace(out)

	return out
}
