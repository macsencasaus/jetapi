package flightradar

import (
	"fmt"
	"log"
	"strings"

	"github.com/macsencasaus/jetapi/scraper"
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

type FlightRadar struct {
	Aircraft     string        `json:"Aircraft"`
	Airline      string        `json:"Airline"`
	Operator     string        `json:"Operator"`
	TypeCode     string        `json:"TypeCode"`
	AirlineCode  string        `json:"AirlineCode"`
	OperatorCode string        `json:"OperatorCode"`
	ModeS        string        `json:"ModeS"`
	Flights      []*flightInfo `json:"Flights"`
}

type FlightRadarRes struct {
	Res *FlightRadar
	Err error
}

const HomeURL = "https://www.flightradar24.com/data/aircraft/"

func GetFlightRadarStruct(reg string, done chan FlightRadarRes) {
	URL := fmt.Sprintf("%s%s", HomeURL, reg)
	b, err := scraper.FetchHTML(URL)
	if err != nil {
		result := FlightRadarRes{Res: nil, Err: err}
		done <- result
		return
	}
	s := scraper.NewScraper(b)

	var aircraft string
	var airline string
	var operator string
	var typeCode string
	var airlineCode string
	var operatorCode string
	var modeS string
	var flights []*flightInfo

	aircraftArr, err := s.FetchText("span", "details", 1)
	if err != nil {
		result := FlightRadarRes{Res: nil, Err: err}
		done <- result
		return
	}
	aircraft = strings.TrimSpace(aircraftArr[0])

	err = s.Advance("span", "details", 1)
	if err != nil {
		result := FlightRadarRes{Res: nil, Err: err}
		done <- result
		return
	}
	airlineArr, err := s.FetchText("a", "", 1)
	if err != nil {
		result := FlightRadarRes{Res: nil, Err: err}
		done <- result
		return
	}
	airline = strings.TrimSpace(airlineArr[0])

	res, err := s.FetchText("span", "details", 5)
	if err != nil {
		result := FlightRadarRes{Res: nil, Err: err}
		done <- result
		return
	}
	operator = strings.TrimSpace(res[0])
	typeCode = strings.TrimSpace(res[1])
	airlineCode = strings.TrimSpace(res[2])
	operatorCode = strings.TrimSpace(res[3])
	modeS = strings.TrimSpace(res[4])

	fr := &FlightRadar{
		Aircraft:     aircraft,
		Airline:      airline,
		Operator:     operator,
		TypeCode:     typeCode,
		AirlineCode:  airlineCode,
		OperatorCode: operatorCode,
		ModeS:        modeS,
		Flights:      flights,
	}

	err = s.Advance("td", "w40 hidden-xs hidden-sm", 3)
	if err != nil {
		log.Printf("flights: %v\n", err)
		result := FlightRadarRes{Res: fr, Err: nil}
		done <- result
		return
	}

	for {
		flight, err := getFlight(s)
		if err != nil {
			if err.Error() == "query not found" {
				break
			}
		}
		flights = append(flights, flight)
	}
	fr.Flights = flights

	result := FlightRadarRes{Res: fr, Err: nil}
	done <- result
}

func getFlight(s *scraper.Scraper) (*flightInfo, error) {
	var date string
	var from string
	var to string
	var flight string
	var flightTime string
	var std string
	var atd string
	var sta string
	var status string

	dateArr, err := s.FetchText("td", "hidden-xs hidden-sm", 1)
	if err != nil {
		return nil, err
	}
	date = strings.TrimSpace(dateArr[0])

	fromToArr, err := s.FetchText("td", "text-center-sm hidden-xs hidden-sm", 2)
	if err != nil {
		return nil, err
	}
	from = strings.TrimSpace(fromToArr[0])
	to = strings.TrimSpace(fromToArr[1])

	err = s.Advance("td", "hidden-xs hidden-sm", 1)
	if err != nil {
		return nil, err
	}

	flightArr, err := s.FetchText("a", "fbold", 1)
	if err != nil {
		return nil, err
	}
	flight = strings.TrimSpace(flightArr[0])

	res, err := s.FetchText("td", "hidden-xs hidden-sm", 4)
	if err != nil {
		return nil, err
	}
	flightTime = strings.TrimSpace(res[0])
	std = strings.TrimSpace(res[1])
	atd = strings.TrimSpace(res[2])
	sta = strings.TrimSpace(res[3])

	statusArr, err := s.FetchText("td", "hidden-xs hidden-sm", 2)
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
