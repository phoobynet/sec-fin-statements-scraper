package scraper

import (
	"github.com/davecgh/go-spew/spew"
	"net/url"
	"testing"
)

func TestBuildLinks(t *testing.T) {
	u, err := url.Parse("https://www.sec.gov/dera/data/financial-statement-data-sets.html")

	if err != nil {
		t.Errorf("failed to parse url: %v", err)
	}

	statementsLinks, buildLinksErr := newStatementLinks(u)

	if buildLinksErr != nil {
		t.Errorf("failed to scrape fin statements links: %v", buildLinksErr)
	}

	spew.Dump(statementsLinks)

}
