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

	log := make(chan string)

	scraper, err := NewFinStatementsScraper(&FinStatementsScraperConfig{
		DatabasePath: dir,
	}, log)

	if err != nil {
		t.Errorf("failed to create scraper: %v", err)
	}

	go func() {
		for {
			select {
			case msg := <-log:
				t.Log(msg)
			}
		}
	}()

	err = scraper.Load(2022, 2)

	if err != nil {
		t.Errorf("failed to load data: %v", err)
	}
}
