package main

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/macsencasaus/jetapi/internal/sites"
)

func (app *application) handleQueries(
	w http.ResponseWriter,
	r *http.Request,
) (*sites.Queries, error) {
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

	q := &sites.Queries{Reg: reg, Photos: photos, Flights: flights}
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
