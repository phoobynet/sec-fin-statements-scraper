package scraper

import (
	"net/url"
	"testing"
)

func TestBuildLinks(t *testing.T) {
	u, err := url.Parse("")

	if err != nil {
		t.Errorf("failed to parse url: %v", err)
	}

	statementsLinks, buildLinksErr := getStatementLinks(u)

	if buildLinksErr != nil {
		t.Errorf("failed to scrape fin statements links: %v", buildLinksErr)
	}

	t.Logf("%+v", statementsLinks)
}
