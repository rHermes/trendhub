package main

import (
	"fmt"
	"log"
	"net/http"
)

func printTableOfLang(tis []TrendingItem) error {
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

	follows := []Language{LangAny, LangGo, LangRust, LangC, LangCPP, LangJava, LangKotlin, LangHaskell}
	for _, f := range follows {
		if err := c.Follow(f); err != nil {
			log.Fatal(err)
		}
	}

	/*
		if err := c.Refresh(); err != nil {
			log.Fatal(err)
		}
	*/

	hh := NewWebsite(c)

	http.ListenAndServe(":8099", hh)
}
