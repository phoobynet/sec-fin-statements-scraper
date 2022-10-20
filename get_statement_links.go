package scraper

import (
	"fmt"
	"github.com/gocolly/colly"
	"log"
	"net/url"
	"strconv"
	"strings"
)

type statementLink struct {
	StatementURL *url.URL
	Title        string
	Year         int
	Quarter      int
	FileName     string
}

type statementLinks struct {
	SourceURL *url.URL
	Links     []statementLink
}

func newStatementLinks(sourceURL *url.URL) (*statementLinks, error) {
	c := colly.NewCollector()

	linkBaseURL := fmt.Sprintf("%s://%s", sourceURL.Scheme, sourceURL.Host)

	links := make([]statementLink, 0)

	c.OnHTML("table.list tbody", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, tr *colly.HTMLElement) {
			link := tr.ChildAttr("a", "href")
			title := tr.ChildText("a")

			fullURLString, err := url.JoinPath(linkBaseURL, link)

			if err != nil {
				log.Fatalln(err)
			}

			statementURL, err := url.Parse(fullURLString)

			if err != nil {
				log.Fatalln(err)
			}

			titleParts := strings.Split(title, " ")

			year, err := strconv.Atoi(titleParts[0])

			if err != nil {
				log.Fatalln(err)
			}

			quarter, err := strconv.Atoi(titleParts[1][1:])

			if err != nil {
				log.Fatalln(err)
			}

			links = append(links, statementLink{
				StatementURL: statementURL,
				Title:        title,
				Year:         year,
				Quarter:      quarter,
				FileName:     parseFileName(statementURL),
			})
		})
	})

	err := c.Visit(sourceURL.String())

	if err != nil {
		return nil, err
	}

	return &statementLinks{
		SourceURL: sourceURL,
		Links:     links,
	}, nil
}

func (s *statementLinks) Get() []statementLink {
	return s.Links
}

func (s *statementLinks) Find(year int, quarter int) *statementLink {
	for _, link := range s.Links {
		if link.Year == year && link.Quarter == quarter {
			return &link
		}
	}

	return nil
}
