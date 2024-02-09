package jetphotos

import (
	"fmt"
	"strings"

	"github.com/macsencasaus/jetapi/scraper"
)

type Jetphoto struct {
	Images        []string `json:"Images"`
	Aircraft      string   `json:"Aircraft"`
	Reg           string   `json:"Reg"`
	Serial        string   `json:"Serial"`
	Airline       string   `json:"Airline"`
	UploadedDates []string `json:"UploadedDates"`
	PhotoDates    []string `json:"PhotoDates"`
	Locations     []string `json:"Locations"`
	Photographers []string `json:"Photographers"`
}

type JetphotoRes struct {
	Res *Jetphoto
	Err error
}

const HomeURL = "https://www.jetphotos.com/"

func GetJetPhotoStruct(reg string, done chan JetphotoRes) {
	URL := fmt.Sprintf("%s/photo/keyword/%s", HomeURL, reg)
	b, err := scraper.FetchHTML(URL)
	if err != nil {
		result := JetphotoRes{Res: nil, Err: err}
		done <- result
		return
	}

	s := scraper.NewScraper(b)
	pageLinks, err := s.FetchLinks("a", "result__photoLink", 3)
	if err != nil {
		result := JetphotoRes{Res: nil, Err: err}
		done <- result
		return
	}
	s.Close()

	imgs := len(pageLinks)
	photoLinks := make([]string, imgs)
	var aircraft, registration, serial, airline string
	uploadedDates := make([]string, imgs)
	photoDates := make([]string, imgs)
	locations := make([]string, imgs)
	photographers := make([]string, imgs)

	for i, link := range pageLinks {
		photoURL := fmt.Sprintf("%s/%s", HomeURL, link)
		b, err := scraper.FetchHTML(photoURL)
		if err != nil {
			result := JetphotoRes{Res: nil, Err: err}
			done <- result
			return
		}

		s := scraper.NewScraper(b)

		// photo links
		photoLinkArr, err := s.FetchLinks("img", "large-photo__img", 1)
		if err != nil {
			result := JetphotoRes{Res: nil, Err: err}
			done <- result
			return
		}
		photoLinks[i] = photoLinkArr[0]

		// registration

		res, err := s.FetchText("h4", "headerText4 color-shark", 3)
		if err != nil {
			result := JetphotoRes{Res: nil, Err: err}
			done <- result
			return
		}
		registration = res[0]
		photoDates[i] = res[1]
		uploadedDates[i] = res[2]

		s.Advance("h2", "header-reset", 1)

		res, err = s.FetchText("a", "link", 3)
		if err != nil {
			result := JetphotoRes{Res: nil, Err: err}
			done <- result
			return
		}
		aircraft = res[0]
		airline = res[1]
		serial = strings.TrimSpace(res[2])

		// location
		s.Advance("h5", "header-reset", 1)
		location, err := s.FetchText("a", "link", 1)
		if err != nil {
			result := JetphotoRes{Res: nil, Err: err}
			done <- result
			return
		}
		locations[i] = location[0]

		// photographer
		photographer, err := s.FetchText("h6", "header-reset", 1)
		if err != nil {
			result := JetphotoRes{Res: nil, Err: err}
			done <- result
			return
		}
		photographers[i] = photographer[0]

		s.Close()
	}

	j := &Jetphoto{
		Images:        photoLinks,
		Aircraft:      aircraft,
		Reg:           registration,
		Serial:        serial,
		Airline:       airline,
		UploadedDates: uploadedDates,
		PhotoDates:    photoDates,
		Locations:     locations,
		Photographers: photographers,
	}

	result := JetphotoRes{Res: j, Err: nil}
	done <- result
}
