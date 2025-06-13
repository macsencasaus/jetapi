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
}

func Scrape(q *APIQueries) (*ScrapeResult, error) {
	var (
		jpResult *JetPhotosResult
		frResult *FlightRadarResult
	)

	g, _ := errgroup.WithContext(context.Background())

	g.Go(func() error {
		res, err := scrapeJetPhotos(q)
		if err != nil {
			return fmt.Errorf("JetPhotos Error: %v", err)
		}
		jpResult = res
		return nil
	})

	g.Go(func() error {
		res, err := scrapeFlightRadar(q)
		if err != nil {
			return fmt.Errorf("FlightRadar Error: %v", err)
		}
		frResult = res
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &ScrapeResult{JetPhotos: jpResult, FlightRadar: frResult}, nil
}
