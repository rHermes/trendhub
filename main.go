package main

import (
	"fmt"
	"log"
)

func printTableOfLang(c *Crawler, lang, period string) error {
	tis, err := c.getTrendingPage(lang, period)
	if err != nil {
		return err
	}
	for i, ti := range tis {
		stars := ti.Stars
		starsToday := ti.StarsToday
		forks := ti.Forks
		lang := ti.Language
		titlelink := ti.RepoOwner + "/" + ti.RepoName
		descr := ti.Description
		pdesc := descr
		if len(pdesc) > 70 {
			pdesc = pdesc[:70] + "..."
		}

		fmt.Printf("%2d: %7d : %5d : %7d : %-10s : %-50s : %s\n", i, stars, starsToday, forks, lang, titlelink, pdesc)
	}
	return nil
}

func main() {
	c, err := NewCrawler("testdir/testdb")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	wewant := [][2]string{
		{LangGo, PeriodDaily},
		{LangHaskell, PeriodDaily},
	}

	for _, w := range wewant {
		fmt.Printf("Printing table for %s on a %s basis.\n", w[0], w[1])

		if err := printTableOfLang(c, w[0], w[1]); err != nil {
			log.Fatal(err)
		}
	}
}
