package scraper

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/gocolly/colly"
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
	url          *url.URL
	databasePath string
}

type FinStatementsScraper struct {
	config *FinStatementsScraperConfig
	links  map[string]*url.URL
}

func NewFinStatementsScraper(config *FinStatementsScraperConfig) (*FinStatementsScraper, error) {
	if config.url == nil {
		u, err := url.Parse("https://www.sec.gov/dera/data/financial-statement-data-sets.html")

		if err != nil {
			return nil, fmt.Errorf("failed to parse url: %w", err)
		}

		config.url = u
	}

	// if the database does not exist, then it will be created
	databaseDir := filepath.Dir(config.databasePath)

	if s, err := os.Stat(databaseDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("databasePath does not exist: %w", err)
	} else if !s.IsDir() {
		return nil, fmt.Errorf("databasePath is not a directory: %w", err)
	}

	links, err := buildLinks(config.url)

	if err != nil {
		return nil, fmt.Errorf("failed to build links: %w", err)
	}

	f := &FinStatementsScraper{
		config: config,
		links:  links,
	}

	return f, nil
}

func (f *FinStatementsScraper) Load(year int, quarter string) error {
	title := fmt.Sprintf("%d %s", year, quarter)

	if link, ok := f.links[title]; !ok {
		f.importFile(link)
	}

	return errors.New("not implemented")
}

func (f *FinStatementsScraper) importFile(url *url.URL) {
	sourceZipFileName := strings.TrimSuffix(getFileName(url), ".zip")
	sourceZipTempFile, err := os.CreateTemp(os.TempDir(), sourceZipFileName)
	defer func(name string) {
		_ = os.Remove(name)
	}(sourceZipTempFile.Name())

	if err != nil {
		log.Fatalln(err)
	}

	zipFileResponse, err := http.Get(url.String())
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(zipFileResponse.Body)

	if err != nil {
		log.Fatalln(err)
	}

	if zipFileResponse.StatusCode == 200 {
		_, err := io.Copy(sourceZipTempFile, zipFileResponse.Body)

		if err != nil {
			log.Fatalln(err)
		}
	}

	zipFile, err := zip.OpenReader(sourceZipTempFile.Name())

	if err != nil {
		log.Fatalln(err)
	}

	for _, fileInZip := range zipFile.File {
		func() {
			zippedFile, fileOpenErr := fileInZip.Open()

			if fileOpenErr != nil {
				log.Fatalln(fileOpenErr)
			}

			txtFile, createTempErr := os.CreateTemp(os.TempDir(), fileInZip.Name)

			if createTempErr != nil {
				log.Fatalln(createTempErr)
			}

			defer func(txtFile *os.File) {
				_ = txtFile.Close()
			}(txtFile)

			_, copyErr := io.Copy(txtFile, zippedFile)

			if copyErr != nil {
				log.Fatalln(copyErr)
			}

			importErr := f.importIntoSQLite(f.config.databasePath, txtFile.Name())

			if importErr != nil {
				log.Fatalln(importErr)
			}
		}()
	}
}

func (f *FinStatementsScraper) importIntoSQLite(dbPath string, txtPath string) error {
	tableName := strings.TrimSuffix(txtPath, ".txt")
	cmd := exec.Command("sqlite3", dbPath, "-tabs", "-cmd", fmt.Sprintf(".import %s %s", txtPath, tableName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func getFileName(u *url.URL) string {
	parts := strings.Split(u.String(), "/")

	return parts[len(parts)-1]
}

func buildLinks(sourceURL *url.URL) (map[string]*url.URL, error) {
	c := colly.NewCollector()

	linkBaseURL := fmt.Sprintf("%s://%s", sourceURL.Scheme, sourceURL.Host)

	links := make(map[string]*url.URL, 0)

	c.OnHTML("table.list tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
			link := tr.ChildAttr("a", "href")
			title := tr.ChildText("a")

			fullURLString, err := url.JoinPath(linkBaseURL, link)

			if err != nil {
				log.Fatalln(err)
			}

			fullURL, err := url.Parse(fullURLString)

			if err != nil {
				log.Fatalln(err)
			}

			links[title] = fullURL
		})
	})

	err := c.Visit(sourceURL.String())

	if err != nil {
		return nil, err
	}

	return links, nil
}
