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
	Links  *statementLinks
	isFile bool
	log    chan string
}

func NewFinStatementsScraper(config *FinStatementsScraperConfig, log chan string) (*FinStatementsScraper, error) {
	if config.SourceURL == nil {
		u, err := url.Parse("https://www.sec.gov/dera/data/financial-statement-data-sets.html")

		if err != nil {
			return nil, fmt.Errorf("failed to parse url: %w", err)
		}

		config.SourceURL = u
	}

	_, statErr := os.Stat(config.DatabasePath)

	if statErr != nil {
		if os.IsNotExist(statErr) {
			_, statErr = os.Stat(filepath.Dir(config.DatabasePath))

			if os.IsNotExist(statErr) {
				return nil, fmt.Errorf("database path does not exist: %w", statErr)
			}
		}
	}

	links, err := newStatementLinks(config.SourceURL)

	if err != nil {
		return nil, fmt.Errorf("failed to build links: %w", err)
	}

	f := &FinStatementsScraper{
		config: config,
		Links:  links,
		isFile: strings.HasSuffix(config.DatabasePath, ".db") || strings.HasSuffix(config.DatabasePath, ".sqlite"),
		log:    log,
	}

	return f, nil
}

func (f *FinStatementsScraper) Load(year int, quarter int) error {
	link := f.Links.Find(year, quarter)

	if link == nil {
		return fmt.Errorf("no link found for year %d and quarter %d", year, quarter)
	}

	f.importFile(link, fmt.Sprintf("%dq%d.db", year, quarter))

	return nil
}

func (f *FinStatementsScraper) LoadLatest() error {
	link := f.Links.Latest()

	if link == nil {
		return fmt.Errorf("no link found for latest")
	}

	f.importFile(link, fmt.Sprintf("%dq%d.db", link.Year, link.Quarter))

	return nil
}

func (f *FinStatementsScraper) importFile(link *statementLink, databaseFileName string) {
	sourceZipFileName := strings.TrimSuffix(link.FileName, ".zip")
	sourceZipTempFile, createTempErr := os.CreateTemp(os.TempDir(), sourceZipFileName)
	defer func(name string) {
		_ = os.Remove(name)
	}(sourceZipTempFile.Name())

	if createTempErr != nil {
		log.Fatalln(createTempErr)
	}

	f.log <- fmt.Sprintf("downloading from %s...", link.StatementURL.String())

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

	f.log <- fmt.Sprintf("downloading from %s...DONE", link.StatementURL.String())

	zipFile, err := zip.OpenReader(sourceZipTempFile.Name())

	if err != nil {
		log.Fatalln(err)
	}

	// for each file, unzip and import into the database
	for _, fileInZip := range zipFile.File {
		if strings.HasSuffix(fileInZip.Name, "readme.htm") {
			continue
		}

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

			var fullDatabasePath string

			if f.isFile {
				fullDatabasePath = f.config.DatabasePath
			} else {
				fullDatabasePath = filepath.Join(f.config.DatabasePath, databaseFileName)
			}

			f.log <- fmt.Sprintf("importing %s into %s...", fileInZip.Name, fullDatabasePath)

			importErr := f.importIntoSQLite(fileInZip.Name, txtTableTempFile.Name(), fullDatabasePath)

			if importErr != nil {
				log.Fatalln(importErr)
			}
		}()
	}
}

func (f *FinStatementsScraper) importIntoSQLite(zipFileName string, txtPath string, fullDatabasePath string) error {
	tableName := strings.TrimSuffix(zipFileName, ".txt")

	cmd := exec.Command("sqlite3", fullDatabasePath, "-tabs", "-cmd", fmt.Sprintf(".import %s %s", txtPath, tableName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
