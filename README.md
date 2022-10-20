# SEC Financial Statements Scraper

Data is source from https://www.sec.gov/dera/data/financial-statement-data-sets.html.

Each `.zip` file contains four `*.txt` files and each file is loaded into a table matching the name.

- `sub.txt` -> `sub`
- `pre.txt` -> `pre`
- `num.txt` -> `num`
- `tag.txt` -> `tag`

The generated database file is called `YYYYqQQ_.zip`, e.g. for year _2022_ and quarter _3_ the generated file would
be `2022q3.zip`.

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