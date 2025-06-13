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
	donejp := make(chan jetPhotosResult)
	donefr := make(chan flightRadarResult)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		getJetPhotosStruct(q, donejp)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		getFlightRadarStruct(q, donefr)
	}()

    go func() {
        wg.Wait()
        close(donejp)
        close(donefr)
    }()

	jp := <-donejp
	fr := <-donefr

	if jp.Err != nil {
		return nil, jp.Err
	}
	if fr.Err != nil {
		return nil, fr.Err
	}

	j := &JetInfo{JetPhotos: jp.Res, FlightRadar: fr.Res}
	return j, nil
}
