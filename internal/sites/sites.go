package sites

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
)

type ScrapeResult struct {
	JetPhotos   *JetPhotosResult
	FlightRadar *FlightRadarResult
}

type APIQueries struct {
	Reg     string
	Photos  int
	Flights int
	OnlyJP  bool
	OnlyFR  bool
}

func Scrape(q *APIQueries) (*ScrapeResult, error) {
	var (
		jpResult *JetPhotosResult
		frResult *FlightRadarResult
	)

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		res, err := ScrapeJetPhotos(q)
		if err != nil {
			return fmt.Errorf("JetPhotos Error: %v", err)
		}
		jpResult = res
		return nil
	})

	g.Go(func() error {
		res, err := ScrapeFlightRadar(q)
		if err != nil {
			return fmt.Errorf("FlightRadar Error: %v", err)
		}
		frResult = res
		return nil
	})

	err := g.Wait()

	if err != nil && jpResult == nil && frResult == nil {
		return nil, err
	}

	return &ScrapeResult{JetPhotos: jpResult, FlightRadar: frResult}, err
}
