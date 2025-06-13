package scraper

import (
	"fmt"
	"strings"
)

type flightInfo struct {
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

type flightRadarInfo struct {
	Aircraft     string        `json:"Aircraft"`
	Airline      string        `json:"Airline"`
	Operator     string        `json:"Operator"`
	TypeCode     string        `json:"TypeCode"`
	AirlineCode  string        `json:"AirlineCode"`
	OperatorCode string        `json:"OperatorCode"`
	ModeS        string        `json:"ModeS"`
	Flights      []*flightInfo `json:"Flights"`
}

type flightRadarRes struct {
	Res *flightRadarInfo
	Err error
}

const frAircraftURL = "https://www.flightradar24.com/data/aircraft/"

func getFlightRadarStruct(q *Queries, done chan flightRadarRes) {
	reg := q.Reg
	URL := fmt.Sprintf("%s%s", frAircraftURL, reg)
	b, err := fetchHTML(URL)
	if err != nil {
		result := flightRadarRes{
			Res: nil,
			Err: frError("fetching fr page", reg, URL, err),
		}
		done <- result
		return
	}

	s := newScraper(b)
    defer s.close()

	var aircraft string
	var airline string
	var operator string
	var typeCode string
	var airlineCode string
	var operatorCode string
	var modeS string
	var flights []*flightInfo

	// aircraft
	aircraftArr, err := s.scrapeText("span", "details", 1)
	if err != nil {
		result := flightRadarRes{
			Res: nil,
			Err: frError("scraping aircraft text", reg, URL, err),
		}
		done <- result
		return
	}
	aircraft = strings.TrimSpace(aircraftArr[0])

	// airline
	err = s.advance("span", "details", 1)
	if err != nil {
		result := flightRadarRes{
			Res: nil,
			Err: frError("advancing to airline text", reg, URL, err),
		}
		done <- result
		return
	}
	airlineArr, err := s.scrapeText("a", "", 1)
	if err != nil {
		result := flightRadarRes{
			Res: nil,
			Err: frError("scraping airline text", reg, URL, err),
		}
		done <- result
		return
	}
	airline = strings.TrimSpace(airlineArr[0])

	// details
	res, err := s.scrapeText("span", "details", 5)
	if err != nil {
		result := flightRadarRes{
			Res: nil,
			Err: frError("scraping details", reg, URL, err),
		}
		done <- result
		return
	}
	operator = strings.TrimSpace(res[0])
	typeCode = strings.TrimSpace(res[1])
	airlineCode = strings.TrimSpace(res[2])
	operatorCode = strings.TrimSpace(res[3])
	modeS = strings.TrimSpace(res[4])

	fr := &flightRadarInfo{
		Aircraft:     aircraft,
		Airline:      airline,
		Operator:     operator,
		TypeCode:     typeCode,
		AirlineCode:  airlineCode,
		OperatorCode: operatorCode,
		ModeS:        modeS,
		Flights:      flights,
	}

	// flights
	err = s.advance("td", "w40 hidden-xs hidden-sm", 3)
	if err != nil {
		result := flightRadarRes{
			Res: fr,
			Err: frError("advancing to flights", reg, URL, err),
		}
		done <- result
		return
	}

	for i := 0; i < q.Flights; i++ {
		flight, err := getFlight(s)
		if err != nil {
			break
		}
		flights = append(flights, flight)
	}
	fr.Flights = flights

	result := flightRadarRes{Res: fr, Err: nil}
	done <- result
}

func getFlight(s *scraper) (*flightInfo, error) {
	var date string
	var from string
	var to string
	var flight string
	var flightTime string
	var std string
	var atd string
	var sta string
	var status string

	// date
	dateArr, err := s.scrapeText("td", "hidden-xs hidden-sm", 1)
	if err != nil {
		return nil, err
	}
	date = strings.TrimSpace(dateArr[0])

	// from & to
	fromToArr, err := s.scrapeText("td", "text-center-sm hidden-xs hidden-sm", 2)
	if err != nil {
		return nil, err
	}
	from = strings.TrimSpace(fromToArr[0])
	to = strings.TrimSpace(fromToArr[1])

	err = s.advance("td", "hidden-xs hidden-sm", 1)
	if err != nil {
		return nil, err
	}

	// flight
	flightArr, err := s.scrapeText("a", "fbold", 1)
	if err != nil {
		return nil, err
	}
	flight = strings.TrimSpace(flightArr[0])

	// time details
	res, err := s.scrapeText("td", "hidden-xs hidden-sm", 4)
	if err != nil {
		return nil, err
	}
	flightTime = strings.TrimSpace(res[0])
	std = strings.TrimSpace(res[1])
	atd = strings.TrimSpace(res[2])
	sta = strings.TrimSpace(res[3])

	// status
	statusArr, err := s.scrapeText("td", "hidden-xs hidden-sm", 2)
	if err != nil {
		return nil, err
	}
	status = strings.TrimSpace(statusArr[1])

	f := &flightInfo{
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
	return fmt.Errorf("Error %s for %s at %s: %s", msg, reg, URL, err.Error())
}
