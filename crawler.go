package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	// We don't want to check for the existance of buckets everywhere, so we make sure
	// they are created at the start
	if err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("following"))
		if err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte("language")); err != nil {
			return err
		}
		return nil
	}); err != nil {
		db.Close()
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

// Following() returns the languages we are following
func (c *Crawler) Following() ([]string, error) {
	var following []string

	if err := c.db.View(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte("following"))

		if err := bk.ForEach(func(k, v []byte) error {
			following = append(following, string(k))
			return nil
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return following, nil
}

func (c *Crawler) Follow(lang string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte("following"))
		return bk.Put([]byte(lang), nil)
	})
}

func (c *Crawler) Unfollow(lang string) error {
	return c.db.Update(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte("following"))
		return bk.Delete([]byte(lang))
	})
}

func (c *Crawler) Refresh() error {
	fs, err := c.Following()
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	for _, f := range fs {
		fmt.Printf("Refreshing language %s\n", f)
		periods := []string{PeriodDaily, PeriodWeekly, PeriodMonthly}
		for _, p := range periods {
			tis, err := c.getTrendingPage(f, p)
			if err != nil {
				return err
			}

			takenAt := time.Now().UTC().Format(time.RFC3339)

			if err := c.db.Update(func(tx *bolt.Tx) error {
				lb := tx.Bucket([]byte("language"))
				llb, err := lb.CreateBucketIfNotExists([]byte(f))
				if err != nil {
					return err
				}
				hlb, err := llb.CreateBucket([]byte(takenAt))
				if err != nil {
					return err
				}

				// We put these into the buckets
				for i, ti := range tis {
					buf.Reset()
					fmt.Fprintf(&buf, "%02d", i)

					j, err := json.Marshal(ti)
					if err != nil {
						return err
					}

					if err := hlb.Put(buf.Bytes(), j); err != nil {
						return err
					}
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}
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
