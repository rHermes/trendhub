// Copyright 2019 Teodor Spæren
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	dbPath = flag.String("db", "testdir/testdb", "the location of the bolt database")
)

func printTableOfLang(tis []TrendingItem) error {
	for i, ti := range tis {
		stars := ti.Stars
		starsToday := ti.StarsIncrease
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

func cmdServeAndRefresh(c *Crawler) error {
	go func(c *Crawler) {
		for {
			if err := c.Refresh(); err != nil {
				log.Printf("[ERR] Couldn't refresh: %s\n", err.Error())
			}
			time.Sleep(4 * time.Hour)
		}
	}(c)
	hh := NewWebsite(c)
	if err := http.ListenAndServe(":8099", hh); err != nil {
		return err
	}
	return nil
}

func Usage() {
	fmt.Println(`use one of the commands:
	follow <lang to follow>+
	follows
	unfollow <lang to unfollow>+
	refresh 
	serve
	serveandrefresh`)
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

	case "serveandrefresh":
		if flag.NArg() != 1 {
			Usage()
		}
		fx = cmdServeAndRefresh
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
