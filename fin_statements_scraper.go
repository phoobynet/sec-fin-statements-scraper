package scraper

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type FinStatementsScraperConfig struct {
	SourceURL    *url.URL // optional - defaults to https://www.sec.gov/dera/data/financial-statement-data-sets.html
	DatabasePath string   // the directory where the database will be created
}

type FinStatementsScraper struct {
	config *FinStatementsScraperConfig
	links  *statementLinks
}

func NewFinStatementsScraper(config *FinStatementsScraperConfig) (*FinStatementsScraper, error) {
	if config.SourceURL == nil {
		u, err := url.Parse("https://www.sec.gov/dera/data/financial-statement-data-sets.html")

		if err != nil {
			return nil, fmt.Errorf("failed to parse url: %w", err)
		}

		config.SourceURL = u
	}

	// if the database does not exist, then it will be created
	databaseDir := filepath.Dir(config.DatabasePath)

	if s, err := os.Stat(databaseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("databasePath does not exist: %w", err)
	} else if !s.IsDir() {
		return nil, fmt.Errorf("databasePath is not a directory: %w", err)
	}

	links, err := newStatementLinks(config.SourceURL)

	if err != nil {
		return nil, fmt.Errorf("failed to build links: %w", err)
	}

	f := &FinStatementsScraper{
		config: config,
		links:  links,
	}

	return f, nil
}

func (f *FinStatementsScraper) Load(year int, quarter int) error {
	link := f.links.Find(year, quarter)

	if link == nil {
		return fmt.Errorf("no link found for year %d and quarter %d", year, quarter)
	}

	f.importFile(link)

	return nil
}

func (f *FinStatementsScraper) importFile(link *statementLink) {
	sourceZipFileName := strings.TrimSuffix(link.FileName, ".zip")
	sourceZipTempFile, createTempErr := os.CreateTemp(os.TempDir(), sourceZipFileName)
	defer func(name string) {
		_ = os.Remove(name)
	}(sourceZipTempFile.Name())

	if createTempErr != nil {
		log.Fatalln(createTempErr)
	}

	zipFileResponse, httpGetErr := http.Get(link.StatementURL.String())
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(zipFileResponse.Body)

	if httpGetErr != nil {
		log.Fatalln(httpGetErr)
	}

	if zipFileResponse.StatusCode == 200 {
		_, copyErr := io.Copy(sourceZipTempFile, zipFileResponse.Body)

		if copyErr != nil {
			log.Fatalln(copyErr)
		}
	}

	zipFile, err := zip.OpenReader(sourceZipTempFile.Name())

	if err != nil {
		log.Fatalln(err)
	}

	// for each file, unzip and import into the database
	for _, fileInZip := range zipFile.File {
		func() {
			zippedFile, fileOpenErr := fileInZip.Open()

			if fileOpenErr != nil {
				log.Fatalln(fileOpenErr)
			}

			txtTableTempFile, createTableTempErr := os.CreateTemp(os.TempDir(), fileInZip.Name)

			if createTableTempErr != nil {
				log.Fatalln(createTableTempErr)
			}

			defer func(txtFile *os.File) {
				_ = txtFile.Close()
				_ = os.Remove(txtFile.Name())
			}(txtTableTempFile)

			_, copyErr := io.Copy(txtTableTempFile, zippedFile)

			if copyErr != nil {
				log.Fatalln(copyErr)
			}

			importErr := f.importIntoSQLite(txtTableTempFile.Name())

			if importErr != nil {
				log.Fatalln(importErr)
			}
		}()
	}
}

func (f *FinStatementsScraper) importIntoSQLite(txtPath string) error {
	tableName := strings.TrimSuffix(txtPath, ".txt")
	cmd := exec.Command("sqlite3", f.config.DatabasePath, "-tabs", "-cmd", fmt.Sprintf(".import %s %s", txtPath, tableName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
