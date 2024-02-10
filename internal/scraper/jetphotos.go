package scraper

import (
	"fmt"
	"strings"
)

type jetPhotosInfo struct {
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

type jetPhotosRes struct {
	Res *jetPhotosInfo
	Err error
}

const jpHomeURL = "https://www.jetphotos.com/"

func getJetPhotosStruct(reg string, done chan jetPhotosRes) {
	URL := fmt.Sprintf("%s/photo/keyword/%s", jpHomeURL, reg)
	b, err := fetchHTML(URL)
	if err != nil {
		result := jetPhotosRes{Res: nil, Err: err}
		done <- result
		return
	}

	s := newScraper(b)
	pageLinks, err := s.fetchLinks("a", "result__photoLink", 3)
	if err != nil {
		result := jetPhotosRes{Res: nil, Err: err}
		done <- result
		return
	}
	s.close()

	imgs := len(pageLinks)
	photoLinks := make([]string, imgs)
	var aircraft, registration, serial, airline string
	uploadedDates := make([]string, imgs)
	photoDates := make([]string, imgs)
	locations := make([]string, imgs)
	photographers := make([]string, imgs)

	for i, link := range pageLinks {
		photoURL := fmt.Sprintf("%s/%s", jpHomeURL, link)
		b, err := fetchHTML(photoURL)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}

		s := newScraper(b)

		// photo links
		photoLinkArr, err := s.fetchLinks("img", "large-photo__img", 1)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		photoLinks[i] = photoLinkArr[0]

		// registration

		res, err := s.fetchText("h4", "headerText4 color-shark", 3)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		registration = res[0]
		photoDates[i] = res[1]
		uploadedDates[i] = res[2]

		s.advance("h2", "header-reset", 1)

		res, err = s.fetchText("a", "link", 3)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		aircraft = res[0]
		airline = res[1]
		serial = strings.TrimSpace(res[2])

		// location
		s.advance("h5", "header-reset", 1)
		location, err := s.fetchText("a", "link", 1)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		locations[i] = location[0]

		// photographer
		photographer, err := s.fetchText("h6", "header-reset", 1)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		photographers[i] = photographer[0]

		s.close()
	}

	j := &jetPhotosInfo{
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

	result := jetPhotosRes{Res: j, Err: nil}
	done <- result
}
