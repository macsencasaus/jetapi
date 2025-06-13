package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/macsencasaus/jetapi/internal/sites"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	page := "home.tmpl"
	app.render(w, http.StatusOK, page, nil)
}

func (app *application) aircraftSearch(w http.ResponseWriter, r *http.Request) {
	page := "aircraft.tmpl"
	q, err := app.parseAPIQueries(w, r)
	if err != nil {
		app.notFoundPage(w)
		return
	}
	q = &sites.APIQueries{Reg: q.Reg, Photos: 3, Flights: 8}
	ji, err := sites.Scrape(q)
	if err != nil {
		app.notFoundPage(w)
		return
	}
	app.render(w, http.StatusOK, page, ji)
}

func (app *application) documentation(w http.ResponseWriter, r *http.Request) {
	page := "documentation.tmpl"
	app.render(w, http.StatusOK, page, nil)
}

func (app *application) queryBuilder(w http.ResponseWriter, r *http.Request) {
	page := "querybuilder.tmpl"
	app.render(w, http.StatusOK, page, nil)
}

func (app *application) notFoundPage(w http.ResponseWriter) {
	page := "notfound.tmpl"
	app.render(w, http.StatusNotFound, page, nil)
}

func (app *application) api(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		app.clientError(w, http.StatusMethodNotAllowed)
		return
	}

	q, err := app.parseAPIQueries(w, r)
	if err != nil {
		app.badRequest(w)
		return
	}

	sr, err := sites.Scrape(q)
	if err != nil {
		app.serverError(w, err)
		return
	}

	jsonResult, err := json.Marshal(sr)
	if err != nil {
		app.serverError(w, fmt.Errorf("Error encoding json: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}
