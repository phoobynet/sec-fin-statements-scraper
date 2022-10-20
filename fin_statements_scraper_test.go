package scraper

import (
	"os"
	"testing"
)

func TestNewFinStatementsScraper(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to get working dir: %v", err)
	}

	scraper, err := NewFinStatementsScraper(&FinStatementsScraperConfig{
		DatabasePath: dir,
	})

	if err != nil {
		t.Errorf("failed to create scraper: %v", err)
	}

	err = scraper.Load(2022, 2)

	if err != nil {
		t.Errorf("failed to load data: %v", err)
	}
}
