package sites

import (
	"encoding/json"
	"fmt"
	"sync"
)

type Result[T any] struct {
	Res T
	Err error
}

type JetInfo struct {
	JetPhotos   *jetPhotosInfo
	FlightRadar *flightRadarInfo
}

type Queries struct {
	Reg     string
	Photos  int
	Flights int
}

func GetJSONData(q *Queries) ([]byte, error) {
	ji, err := GetJetInfo(q)
	if err != nil {
		return nil, fmt.Errorf("Jetphotos Error: %v", err)
	}
	jsonData, err := json.Marshal(ji)
	if err != nil {
		return nil, fmt.Errorf("FlightRadar Error: %v", err)
	}

	return jsonData, nil
}

func GetJetInfo(q *Queries) (*JetInfo, error) {
	jpCh := make(chan jetPhotosResult)
	frCh := make(chan flightRadarResult)

	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		jpCh <- getJetPhotosStruct(q)
	}()

	go func() {
		defer wg.Done()
		frCh <- getFlightRadarStruct(q)
	}()

	go func() {
		wg.Wait()
		close(jpCh)
		close(frCh)
	}()

	jp := <-jpCh
	fr := <-frCh

	if jp.Err != nil {
		return nil, jp.Err
	}
	if fr.Err != nil {
		return nil, fr.Err
	}

	j := &JetInfo{JetPhotos: jp.Res, FlightRadar: fr.Res}
	return j, nil
}
