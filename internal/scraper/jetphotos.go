package scraper

import (
	"fmt"
	"strings"
)

type jetPhotosInfo struct {
	Reg    string         `json:"Reg"`
	Images []imagesStruct `json:"Images"`
}

type imagesStruct struct {
	Image        string          `json:"Image"`
	Link         string          `json:"Link"`
	Thumbnail    string          `json:"Thumbnail"`
	DateTaken    string          `json:"DateTaken"`
	DateUploaded string          `json:"DateUploaded"`
	Location     string          `json:"Location"`
	Photographer string          `json:"Photographer"`
	Aircraft     *aircraftStruct `json:"Aircraft"`
}

type aircraftStruct struct {
	Aircraft string `json:"Aircraft"`
	Serial   string `json:"Serial"`
	Airline  string `json:"Airline"`
}

type jetPhotosRes struct {
	Res *jetPhotosInfo
	Err error
}

const jpHomeURL = "https://www.jetphotos.com"

func getJetPhotosStruct(q *Queries, done chan jetPhotosRes) {
	if q.Photos == 0 {
		result := jetPhotosRes{Res: &jetPhotosInfo{Reg: strings.ToUpper(q.Reg)}}
		done <- result
		return
	}

	URL := fmt.Sprintf("%s/photo/keyword/%s", jpHomeURL, q.Reg)
	b, err := fetchHTML(URL)
	if err != nil {
		result := jetPhotosRes{Res: nil, Err: err}
		done <- result
		return
	}

	s := newScraper(b)
	pageLinks := []string{}
	thumbnails := []string{}
	atLeastOne := false
	for i := 0; i < q.Photos; i++ {
		pageLink, err1 := s.fetchLinks("a", "result__photoLink", 1)
		thumbnail, err2 := s.fetchLinks("img", "result__photo", 1)
		if err1 != nil || err2 != nil {
			if atLeastOne {
				break
			}
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		pageLinks = append(pageLinks, pageLink[0])
		thumbnails = append(thumbnails, thumbnail[0])
		atLeastOne = true
	}
	s.close()

	imgs := len(pageLinks)

	var registration string
	images := make([]imagesStruct, imgs)

	for i, link := range pageLinks {
		photoURL := fmt.Sprintf("%s%s", jpHomeURL, link)
		images[i].Link = photoURL
		images[i].Thumbnail = "https:" + thumbnails[i]

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
		images[i].Image = photoLinkArr[0]

		// registration
		res, err := s.fetchText("h4", "headerText4 color-shark", 3)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		registration = res[0]
		images[i].DateTaken = res[1]
		images[i].DateUploaded = res[2]

		s.advance("h2", "header-reset", 1)

		aircraft := &aircraftStruct{}
		res, err = s.fetchText("a", "link", 3)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		aircraft.Aircraft = res[0]
		aircraft.Airline = res[1]
		aircraft.Serial = strings.TrimSpace(res[2])
		images[i].Aircraft = aircraft

		// location
		s.advance("h5", "header-reset", 1)
		location, err := s.fetchText("a", "link", 1)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		images[i].Location = location[0]

		// photographer
		photographer, err := s.fetchText("h6", "header-reset", 1)
		if err != nil {
			result := jetPhotosRes{Res: nil, Err: err}
			done <- result
			return
		}
		images[i].Photographer = photographer[0]

		s.close()
	}

	j := &jetPhotosInfo{
		Images: images,
		Reg:    registration,
	}

	result := jetPhotosRes{Res: j, Err: nil}
	done <- result
}
