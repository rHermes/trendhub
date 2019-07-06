package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	bolt "go.etcd.io/bbolt"
)

type Crawler struct {
	c  http.Client
	db *bolt.DB
}

const (
	LangGo      = "go"
	LangC       = "c"
	LangCPP     = "c++"
	LangRust    = "rust"
	LangHaskell = "haskell"

	PeriodDaily   = "daily"
	PeriodWeekly  = "weekly"
	PeriodMonthly = "monthly"
)

// NewCrawler Returns a new crawler
func NewCrawler(dbpath string) (*Crawler, error) {
	db, err := bolt.Open(dbpath, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Crawler{db: db}, nil
}

// Close closes the crawler
func (c *Crawler) Close() error {
	return c.db.Close()
}

func (c *Crawler) getTrendingPage(lang string, period string) ([]TrendingItem, error) {
	u := fmt.Sprintf("https://github.com/trending/%s?since=%s", lang, period)
	res, err := c.c.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return parsePage(res.Body)
}

func (c *Crawler) Refresh() error {
	res, err := c.c.Get("https://github.com/trending")
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return nil
}

type TrendingItem struct {
	RepoOwner   string
	RepoName    string
	Description string
	Language    string
	Forks       int
	Stars       int
	StarsToday  int
}

func parsePage(body io.Reader) ([]TrendingItem, error) {
	doc, err := goquery.NewDocumentFromReader(body)
	if err != nil {
		return nil, err
	}

	items := make([]TrendingItem, 0)
	doc.Find("article.Box-row").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Repolink, repo organization and repo name
		titlelink, ok := s.Find("h1.h3.lh-condensed > a").Attr("href")
		if !ok {
			err = fmt.Errorf("%d: Couldn't get titlelink: %s", i, err.Error())
			return false
		}
		pars := strings.Split(titlelink, "/")
		if len(pars) != 3 {
			err = fmt.Errorf("%d: There was more than 3 items in the titlelink split!", i)
			return false
		}
		repoOwner, repoName := pars[1], pars[2]

		// Description
		q := s.Find("p")
		if q.Length() > 1 {
			kk, _ := s.Html()
			err = fmt.Errorf("%d: We expect only one <p> tag! %s", i, kk)
			return false
		}
		descr := strings.TrimSpace(q.Text())

		// Programming language
		q = s.Find(`span[itemprop="programmingLanguage"]`)
		lang := strings.TrimSpace(q.Text())
		if q.Length() == 0 {
			lang = "Unknown"
		} else if q.Length() != 1 {
			err = fmt.Errorf("%d: We expect only one programmingLanguage span!", i)
			return false
		}

		// Stargazers
		q = s.Find(fmt.Sprintf(`a[href="%s/stargazers"]`, titlelink))
		if q.Length() != 1 {
			err = fmt.Errorf("%d: We expected exactly one stargazers link", i)
			return false
		}
		starsRaw := strings.ReplaceAll(strings.TrimSpace(q.Text()), ",", "")
		stars, err := strconv.Atoi(starsRaw)
		if err != nil {
			err = fmt.Errorf("%d: We couldn't convert starsRaw to stars: %s", i, err.Error())
			return false
		}

		// forks
		q = s.Find(fmt.Sprintf(`a[href="%s/network/members"]`, titlelink))
		if q.Length() != 1 {
			err = fmt.Errorf("%d: We expect exactly one members link", i)
			return false
		}
		forksRaw := strings.ReplaceAll(strings.TrimSpace(q.Text()), ",", "")
		forks, err := strconv.Atoi(forksRaw)
		if err != nil {
			err = fmt.Errorf("%d: We couldn't convert forksRaw to forks: %s", i, err.Error())
			return false
		}

		q = s.Find("span.float-sm-right")
		if q.Length() != 1 {
			err = fmt.Errorf("%d: We expected exactly one stars today object", i)
			return false
		}
		starsTodayRaw := strings.ReplaceAll(strings.TrimSuffix(strings.TrimSpace(q.Text()), " stars today"), ",", "")
		starsToday, err := strconv.Atoi(starsTodayRaw)
		if err != nil {
			err = fmt.Errorf("%d: We couldn't convert starsTodayRaw to starsToday: %s", i, err.Error())
			return false
		}

		ti := TrendingItem{
			RepoOwner:   repoOwner,
			RepoName:    repoName,
			Stars:       stars,
			StarsToday:  starsToday,
			Forks:       forks,
			Description: descr,
			Language:    lang,
		}
		items = append(items, ti)

		return true
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}
