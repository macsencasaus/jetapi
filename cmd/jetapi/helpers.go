package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/macsencasaus/jetapi/internal/sites"
)

func (app *application) logErr(err error) {
	trace := fmt.Sprintf("%v\n%s", err, debug.Stack())
	app.errorLog.Println(trace)
}

func (app *application) serverError(w http.ResponseWriter, err error) {
	app.logErr(err)
	status := http.StatusInternalServerError
	http.Error(w, http.StatusText(status), status)
}

func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *application) notFound(w http.ResponseWriter) {
	app.clientError(w, http.StatusNotFound)
}

func (app *application) badRequest(w http.ResponseWriter) {
	app.clientError(w, http.StatusBadRequest)
}

func (app *application) render(
	w http.ResponseWriter,
	status int,
	page string,
	data *sites.ScrapeResult,
) {
	ts, ok := app.templateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.serverError(w, err)
		return
	}

	w.WriteHeader(status)

	err := ts.ExecuteTemplate(w, "base", data)
	if err != nil {
		app.serverError(w, err)
	}
}

func (app *application) parseAPIQueries(
	w http.ResponseWriter,
	r *http.Request,
) (*sites.APIQueries, error) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		return nil, fmt.Errorf("%d", http.StatusMethodNotAllowed)
	}
	queryParams := r.URL.Query()

	reg := queryParams.Get("reg")
	if reg == "" {
		return nil, fmt.Errorf("%d", http.StatusNotFound)
	}

	// is alphanumeric and may include '-'
	isAlpha := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(reg)
	if !isAlpha {
		return nil, fmt.Errorf("%d", http.StatusBadRequest)
	}

	onlyJP := queryParams.Get("only_jp") == "true"
	onlyFR := queryParams.Get("only_fr") == "true"

	photos, err := handleNumQuery(queryParams, "photos")
	if err != nil {
		return nil, err
	}
	if photos == -1 {
		photos = 3
	}

	flights, err := handleNumQuery(queryParams, "flights")
	if err != nil {
		return nil, err
	}
	if flights == -1 {
		flights = 20
	}

	q := &sites.APIQueries{
		Reg:     reg,
		Photos:  photos,
		Flights: flights,
		OnlyJP:  onlyJP,
		OnlyFR:  onlyFR,
	}
	return q, nil
}

func handleNumQuery(qp url.Values, query string) (int, error) {
	resStr := qp.Get(query)
	if resStr == "" {
		return -1, nil
	}
	res, err := strconv.Atoi(resStr)
	if err != nil || res < 0 {
		return 0, fmt.Errorf("%d", http.StatusBadRequest)
	}
	return res, nil
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	pages, err := filepath.Glob("./ui/html/pages/*.tmpl.html")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		ts, err := template.ParseFiles("./ui/html/base.tmpl.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob("./ui/html/partial/*.tmpl.html")
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func (app *application) statsLogger() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		app.statsMu.Lock()

		var avgLatency time.Duration
		if app.apiCalls > 0 {
			avgLatency = app.totalLatency / time.Duration(app.apiCalls)
		}

		msg := `
+-----------------------------+
|        API Statistics       |
+-----------------------------+
| Calls to /api     : %7d |
| Average Latency   : %7s |
+-----------------------------+`

		app.infoLog.Printf(msg, app.apiCalls, avgLatency.Round(time.Millisecond))

		app.apiCalls = 0
		app.totalLatency = 0

		app.statsMu.Unlock()
	}
}
