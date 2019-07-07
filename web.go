package main

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

const (
	ctxCrawler = "__crawler__"
	ctxIdxTmpl = "__indexTemplate__"
)

func indexPage(w http.ResponseWriter, r *http.Request) {
	tmpl := r.Context().Value(ctxIdxTmpl).(*template.Template)
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, nil)
	if err != nil {
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
	lt := template.Must(template.ParseFiles("templates/layout.tmpl.html"))
	indexTemplate := template.Must(lt.ParseFiles("templates/index.tmpl.html"))

	r.Use(middleware.WithValue(ctxIdxTmpl, indexTemplate))

	r.Get("/", indexPage)

	return r
}
