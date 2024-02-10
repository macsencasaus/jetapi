package main

import (
	"net/http"
	"regexp"

	"github.com/macsencasaus/jetapi/internal/scraper"
)

func (app *application) api(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	queryParams := r.URL.Query()

	reg := queryParams.Get("reg")
	if reg == "" {
		http.Error(w, "Invalid parameter: reg is required", http.StatusBadRequest)
		return
	}

	// is alphanumeric and may include '-'
	isAlpha := regexp.MustCompile(`^[a-zA-Z0-9-]+$`).MatchString(reg)
	if !isAlpha {
		http.Error(
			w,
			"Invalid parameter: reg must be alphanumeric allowing for '-'",
			http.StatusBadRequest,
		)
		return
	}

	jsonData, err := scraper.GetJSONData(reg)
	if err != nil {
		app.errorLog.Println(err.Error())
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}