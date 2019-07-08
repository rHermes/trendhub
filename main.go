package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	dbPath = flag.String("db", "testdir/testdb", "the location of the bolt database")
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

func cmdFollows(c *Crawler) error {
	fs, err := c.Follows()
	if err != nil {
		return err
	}
	for _, f := range fs {
		fmt.Println(f.StoreName)
	}
	return nil
}

func cmdServe(c *Crawler) error {
	hh := NewWebsite(c)
	if err := http.ListenAndServe(":8099", hh); err != nil {
		return err
	}
	return nil
}

func cmdFollow(c *Crawler) error {
	var ls []Language
	for i := 1; i < flag.NArg(); i++ {
		l, ok := StoreToLang[flag.Arg(i)]
		if !ok {
			return fmt.Errorf("Unknown language: %s", flag.Arg(i))
		}
		ls = append(ls, l)
	}
	for _, l := range ls {
		if err := c.Follow(l); err != nil {
			return err
		}
	}
	return nil
}

func cmdUnfollow(c *Crawler) error {
	var ls []Language
	for i := 1; i < flag.NArg(); i++ {
		l, ok := StoreToLang[flag.Arg(i)]
		if !ok {
			return fmt.Errorf("Unknown language: %s", flag.Arg(i))
		}
		ls = append(ls, l)
	}
	for _, l := range ls {
		if err := c.Unfollow(l); err != nil {
			return err
		}
	}
	return nil
}

func cmdRefresh(c *Crawler) error {
	return c.Refresh()
}

func Usage() {
	fmt.Println(`use one of the commands:
	follow <lang to follow>+
	follows
	unfollow <lang to unfollow>+
	refresh 
	serve`)
	os.Exit(1)
}

func main() {
	flag.Parse()

	var fx func(c *Crawler) error
	var err error

	switch strings.ToLower(flag.Arg(0)) {
	case "follows":
		if flag.NArg() != 1 {
			Usage()
		}
		fx = cmdFollows
	case "follow":
		if flag.NArg() < 2 {
			Usage()
		}
		fx = cmdFollow
	case "unfollow":
		if flag.NArg() < 2 {
			Usage()
		}
		fx = cmdUnfollow
	case "serve":
		if flag.NArg() != 1 {
			Usage()
		}
		fx = cmdServe
	case "refresh":
		if flag.NArg() != 1 {
			Usage()
		}
		fx = cmdRefresh
	default:
		Usage()
	}

	c, err := NewCrawler(*dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	if err := fx(c); err != nil {
		log.Fatal(err)
	}
}
