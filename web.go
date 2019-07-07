package main

import (
	"bytes"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	ctxCrawler = "__crawler__"
	ctxIdxTmpl = "__indexTemplate__"
)

type LanguageScrape struct {
	Lang    Language
	Items   []TrendingItem
	Scraped time.Time
}
type IndexPageCtx struct {
	Period  string
	Langs   []LanguageScrape
	BoltDur time.Duration
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	c := r.Context().Value(ctxCrawler).(*Crawler)

	var pctx IndexPageCtx
	pctx.Period = PeriodDaily

	tStart := time.Now()

	fs, err := c.Following()
	if err != nil {
		http.Error(w, "Some error with loading: "+err.Error(), http.StatusInternalServerError)
		return
	}

	for _, f := range fs {
		tis, ts, err := c.Latest(f, pctx.Period)
		if err != nil {
			// TODO(rHermes): Create some kind of blank page when we have no scrape?
			if err == ErrNoScrapesForLang || err == ErrNoScrapesForPeriod {
				continue
			}
			http.Error(w, "Some error with loading: "+err.Error(), http.StatusInternalServerError)
			return
		}
		pctx.Langs = append(pctx.Langs, LanguageScrape{
			Lang:    f,
			Items:   tis,
			Scraped: ts,
		})
	}
	pctx.BoltDur = time.Since(tStart)

	tmpl := r.Context().Value(ctxIdxTmpl).(*template.Template)
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, pctx); err != nil {
		http.Error(w, "Some error with templates: "+err.Error(), http.StatusInternalServerError)
		return
	}
	buf.WriteTo(w)
}

func NewWebsite(c *Crawler) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(middleware.WithValue(ctxCrawler, c))

	// Here we create the templates
	lt := template.Must(template.ParseFiles(
		"templates/layout.html.tmpl",
		"templates/trending-lang.html.tmpl",
		"templates/trending-item.html.tmpl",
	))
	indexTemplate := template.Must(lt.ParseFiles("templates/index.html.tmpl"))

	r.Use(middleware.WithValue(ctxIdxTmpl, indexTemplate))

	workDir, _ := os.Getwd()
	staticDir := filepath.Join(workDir, "static")
	r.Get("/", indexPage)
	FileServer(r, "/static", http.Dir(staticDir))

	return r
}

// FileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
