package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/macsencasaus/jetapi/internal/sites"
)

type handlerFunc = func(http.ResponseWriter, *http.Request)

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()

	fileServer := http.FileServer(http.Dir("./ui/static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("/", app.home)
	mux.HandleFunc("/api", app.statsMiddleware(app.api))
	mux.HandleFunc("/aircraft", app.aircraftSearch)
	mux.HandleFunc("/documentation", app.documentation)
	mux.HandleFunc("/querybuilder", app.queryBuilder)

	return mux
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}
	page := "home.tmpl.html"
	app.render(w, http.StatusOK, page, nil)
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

	var jsonResult []byte

	if q.OnlyJP == q.OnlyFR {
		sr, err := sites.Scrape(q)
		if sr == nil {
			app.notFound(w)
			return
		}

		if err != nil {
			app.logErr(fmt.Errorf("Partial Error: %v", err))
		}

		jsonResult, err = json.Marshal(sr)
	} else if q.OnlyJP {
		jpRes, err := sites.ScrapeJetPhotos(q)
		if err != nil {
			app.notFound(w)
			return
		}

		jsonResult, err = json.Marshal(jpRes)
	} else if q.OnlyFR {
		frRes, err := sites.ScrapeFlightRadar(q)
		if err != nil {
			app.notFound(w)
			return
		}

		jsonResult, err = json.Marshal(frRes)
	}

	if err != nil {
		app.serverError(w, fmt.Errorf("Error encoding json: %v", err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
}

func (app *application) aircraftSearch(w http.ResponseWriter, r *http.Request) {
	page := "aircraft.tmpl.html"
	q, err := app.parseAPIQueries(w, r)
	if err != nil {
		app.notFoundPage(w)
		return
	}
	q = &sites.APIQueries{Reg: q.Reg, Photos: 3, Flights: 8}
	sr, err := sites.Scrape(q)
	if sr == nil {
		app.notFoundPage(w)
		return
	}
	app.render(w, http.StatusOK, page, sr)
}

func (app *application) documentation(w http.ResponseWriter, r *http.Request) {
	page := "documentation.tmpl.html"
	app.render(w, http.StatusOK, page, nil)
}

func (app *application) queryBuilder(w http.ResponseWriter, r *http.Request) {
	page := "querybuilder.tmpl.html"
	app.render(w, http.StatusOK, page, nil)
}

func (app *application) notFoundPage(w http.ResponseWriter) {
	page := "notfound.tmpl.html"
	app.render(w, http.StatusNotFound, page, nil)
}

func (app *application) statsMiddleware(next handlerFunc) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		next(w, r)
		latency := time.Since(now)

		app.statsMu.Lock()
		defer app.statsMu.Unlock()

		app.apiCalls++
		app.totalLatency += latency
	}
}
