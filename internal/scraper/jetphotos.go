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

	reg := q.Reg
	URL := fmt.Sprintf("%s/photo/keyword/%s", jpHomeURL, reg)
	b, err := fetchHTML(URL)
	if err != nil {
		result := jetPhotosRes{
			Res: nil,
			Err: jpError("scraping search URL", reg, URL, err),
		}
		done <- result
		return
	}

	s := newScraper(b)
    defer s.close()

	pageLinks := []string{}
	thumbnails := []string{}
	for i := 0; i < q.Photos; i++ {
		pageLink, err := s.scrapeLinks("a", "result__photoLink", 1)
		if err != nil {
			if len(pageLinks) > 0 {
				break
			}
			result := jetPhotosRes{
				Res: nil,
				Err: jpError("scraping aircraft pagelinks", reg, URL, err),
			}
			done <- result
			return
		}

		thumbnail, err := s.scrapeLinks("img", "result__photo", 1)
		if err != nil {
			if len(thumbnails) > 0 {
				break
			}
			result := jetPhotosRes{
				Res: nil,
				Err: jpError("scraping aircraft thumbnails", reg, URL, err),
			}
			done <- result
			return
		}
		pageLinks = append(pageLinks, pageLink[0])
		thumbnails = append(thumbnails, thumbnail[0])
	}

	imgs := len(pageLinks)

	var registration string
	images := make([]imagesStruct, imgs)

	for i, link := range pageLinks {
		photoURL := fmt.Sprintf("%s%s", jpHomeURL, link)
		images[i].Link = photoURL
		images[i].Thumbnail = "https:" + thumbnails[i]

		b, err := fetchHTML(photoURL)
		if err != nil {
			result := jetPhotosRes{
				Res: nil,
				Err: jpError("fetching HTML page", reg, URL, err),
			}
			done <- result
			return
		}

		s := newScraper(b)
        defer s.close()

		// photo links
		photoLinkArr, err := s.scrapeLinks("img", "large-photo__img", 1)
		if err != nil {
			result := jetPhotosRes{
				Res: nil,
				Err: jpError("scraping photo links", reg, URL, err),
			}
			done <- result
			return
		}
		images[i].Image = photoLinkArr[0]

		// registration
		res, err := s.scrapeText("h4", "headerText4 color-shark", 3)
		if err != nil {
			result := jetPhotosRes{
				Res: nil,
				Err: jpError("scraping registrating text", reg, URL, err),
			}
			done <- result
			return
		}
		registration = res[0]
		images[i].DateTaken = res[1]
		images[i].DateUploaded = res[2]

		// aircraft
		s.advance("h2", "header-reset", 1)
		aircraft := &aircraftStruct{}
		res, err = s.scrapeText("a", "link", 3)
		if err != nil {
			result := jetPhotosRes{
				Res: nil,
				Err: jpError("scraping aircraft text", reg, URL, err),
			}
			done <- result
			return
		}
		aircraft.Aircraft = res[0]
		aircraft.Airline = res[1]
		aircraft.Serial = strings.TrimSpace(res[2])
		images[i].Aircraft = aircraft

		// location
		s.advance("h5", "header-reset", 1)
		location, err := s.scrapeText("a", "link", 1)
		if err != nil {
			result := jetPhotosRes{
				Res: nil,
				Err: jpError("scraping location text", reg, URL, err),
			}
			done <- result
			return
		}
		images[i].Location = location[0]

		// photographer
		photographer, err := s.scrapeText("h6", "header-reset", 1)
		if err != nil {
			result := jetPhotosRes{
				Res: nil,
				Err: jpError("scraping photographer text", reg, URL, err),
			}
			done <- result
			return
		}
		images[i].Photographer = photographer[0]
	}

	j := &jetPhotosInfo{
		Images: images,
		Reg:    registration,
	}

	result := jetPhotosRes{Res: j, Err: nil}
	done <- result
}

func jpError(msg, reg, url string, err error) error {
	return fmt.Errorf("Error %s for %s at %s: %s", msg, reg, url, err.Error())
}
