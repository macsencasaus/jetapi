package api

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"sync"

	"github.com/macsencasaus/jetapi/flightradar"
	"github.com/macsencasaus/jetapi/jetphotos"
)

type jetInfo struct {
	JetPhotos   *jetphotos.Jetphoto
	FlightRadar *flightradar.FlightRadar
}

func getJSONData(reg string) ([]byte, error) {
	donejp := make(chan jetphotos.JetphotoRes)
	donefr := make(chan flightradar.FlightRadarRes)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		jetphotos.GetJetPhotoStruct(reg, donejp)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		flightradar.GetFlightRadarStruct(reg, donefr)
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

	j := jetInfo{JetPhotos: jp.Res, FlightRadar: fr.Res}
	jsonData, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func api(w http.ResponseWriter, r *http.Request) {
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

	jsonData, err := getJSONData(reg)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", api)

	log.Println("Server listenining on :8080...")
	err := http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}
