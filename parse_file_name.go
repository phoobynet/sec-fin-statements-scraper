package scraper

import (
	"net/url"
	"strings"
)

func parseFileName(u *url.URL) string {
	parts := strings.Split(u.String(), "/")

	return parts[len(parts)-1]
}
