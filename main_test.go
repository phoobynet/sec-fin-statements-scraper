package scraper

import "testing"

func TestFoo(t *testing.T) {
	scraper, err := NewFinStatementsScraper(&FinStatementsScraperConfig{
		databasePath: "/Volumes/yotta_1tb/sec_data",
	})

	if err != nil {
		return
	}

	statementsLinks := scraper.BuildLinks()

	t.Logf("%+v", statementsLinks)
}
