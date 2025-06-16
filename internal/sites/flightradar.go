package sites

import (
	"fmt"
	"strings"

	"github.com/macsencasaus/jetapi/internal/scraper"
)

type FlightRadarResult struct {
	Aircraft     string              `json:"Aircraft"`
	Airline      string              `json:"Airline"`
	Operator     string              `json:"Operator"`
	TypeCode     string              `json:"TypeCode"`
	AirlineCode  string              `json:"AirlineCode"`
	OperatorCode string              `json:"OperatorCode"`
	ModeS        string              `json:"ModeS"`
	Flights      []*FlightAttributes `json:"Flights"`
}

type FlightAttributes struct {
	Date       string `json:"Date"`
	From       string `json:"From"`
	To         string `json:"To"`
	Flight     string `json:"Flight"`
	FlightTime string `json:"FlightTime"`
	STD        string `json:"STD"`
	ATD        string `json:"ATD"`
	STA        string `json:"STA"`
	Status     string `json:"Status"`
}

const frAircraftURL = "https://www.flightradar24.com/data/aircraft/"

func scrapeFlightRadar(q *APIQueries) (*FlightRadarResult, error) {
	reg := q.Reg
	URL := fmt.Sprintf("%s%s", frAircraftURL, reg)
	b, err := scraper.FetchHTML(URL)
	if err != nil {
		return nil, frError("fetching fr page", reg, URL, err)
	}

	s := scraper.NewScraper(b)
	defer s.Close()

	// aircraft
	aircraftArr, err := s.ScrapeText("span", "details", 1)
	if err != nil {
		return nil, frError("scraping aircraft text", reg, URL, err)
	}
	aircraft := strings.TrimSpace(aircraftArr[0])

	// airline
	// can either be a link, typically for commerical airlines,
	// or text, for private owners
	err = s.Advance("span", "details", 1)
	if err != nil {
		return nil, frError("advancing to airline text", reg, URL, err)
	}
	airline, ok := s.TryScrapeText()
	if !ok {
		airlineArr, err := s.ScrapeText("a", "", 1)
		if err != nil {
			return nil, frError("scraping airline text", reg, URL, err)
		}
		airline = strings.TrimSpace(airlineArr[0])
	} else {
		airline = strings.TrimSpace(airline)
	}

	// details
	res, err := s.ScrapeText("span", "details", 5)
	if err != nil || len(res) != 5 {
		fmt.Printf("details: %v\n", res)
		return nil, frError("scraping details", reg, URL, err)
	}
	operator := strings.TrimSpace(res[0])
	typeCode := strings.TrimSpace(res[1])
	airlineCode := strings.TrimSpace(res[2])
	operatorCode := strings.TrimSpace(res[3])
	modeS := strings.TrimSpace(res[4])

	response := &FlightRadarResult{
		Aircraft:     aircraft,
		Airline:      airline,
		Operator:     operator,
		TypeCode:     typeCode,
		AirlineCode:  airlineCode,
		OperatorCode: operatorCode,
		ModeS:        modeS,
		Flights:      []*FlightAttributes{},
	}

	// flights
	err = s.Advance("td", "w40 hidden-xs hidden-sm", 3)
	if err != nil {
		// doesn't find any flights
		return response, nil
	}

	for i := 0; i < q.Flights; i++ {
		flight, err := scrapeFlight(s)
		if err != nil {
			break
		}
		response.Flights = append(response.Flights, flight)
	}

	return response, nil
}

func scrapeFlight(s *scraper.Scraper) (*FlightAttributes, error) {
	// date
	dateArr, err := s.ScrapeText("td", "hidden-xs hidden-sm", 1)
	if err != nil {
		return nil, err
	}
	date := strings.TrimSpace(dateArr[0])

	// from & to
	fromToArr, err := s.ScrapeText("td", "text-center-sm hidden-xs hidden-sm", 2)
	if err != nil {
		return nil, err
	}
	from := strings.TrimSpace(fromToArr[0])
	to := strings.TrimSpace(fromToArr[1])

	err = s.Advance("td", "hidden-xs hidden-sm", 1)
	if err != nil {
		return nil, err
	}

	// flight
	flightArr, err := s.ScrapeText("a", "fbold", 1)
	if err != nil {
		return nil, err
	}
	flight := strings.TrimSpace(flightArr[0])

	// time details
	res, err := s.ScrapeText("td", "hidden-xs hidden-sm", 4)
	if err != nil {
		return nil, err
	}
	flightTime := strings.TrimSpace(res[0])
	std := strings.TrimSpace(res[1])
	atd := strings.TrimSpace(res[2])
	sta := strings.TrimSpace(res[3])

	// status
	statusArr, err := s.ScrapeText("td", "hidden-xs hidden-sm", 2)
	if err != nil {
		return nil, err
	}
	status := strings.TrimSpace(statusArr[1])

	f := &FlightAttributes{
		Date:       date,
		From:       from,
		To:         to,
		Flight:     flight,
		FlightTime: flightTime,
		STD:        std,
		ATD:        atd,
		STA:        sta,
		Status:     status,
	}
	return f, nil
}

func frError(msg, reg, URL string, err error) error {
	return fmt.Errorf("Error %s for %s at %s: %v", msg, reg, URL, err)
}
