# SEC Financial Statements Scraper

```go
import "github.com/phoobynet/sec-fin-statements-scraper"

func main () {
    sc, err := scraper.NewFinStatementsScraper(&FinStatementsScraperConfig{
        // this is where the database to end up
        DatabasePath: "~",
    })

    if err != nil {
        log.Fatal(err)
    }
    
	// file dumped to ~/2022q2.db
    err := sc.Load(2022, 2)
	
    if err != nil {
        log.Fatal(err)	
    }
}
```