package sites

import (
	"fmt"
	"strings"

	"github.com/macsencasaus/jetapi/internal/scraper"
)

type JetPhotosResult struct {
	Reg    string            `json:"Reg"`
	Images []ImageAttributes `json:"Images"`
}

type ImageAttributes struct {
	Image        string `json:"Image"`
	Link         string `json:"Link"`
	Thumbnail    string `json:"Thumbnail"`
	DateTaken    string `json:"DateTaken"`
	DateUploaded string `json:"DateUploaded"`
	Location     string `json:"Location"`
	Photographer string `json:"Photographer"`
	Aircraft     string `json:"Aircraft"`
	Serial       string `json:"Serial"`
	Airline      string `json:"Airline"`
}

const jpHomeURL = "https://www.jetphotos.com"

func scrapeJetPhotos(q *APIQueries) (*JetPhotosResult, error) {
	reg := q.Reg
	if q.Photos == 0 {
		return &JetPhotosResult{Reg: strings.ToUpper(reg)}, nil
	}

	URL := fmt.Sprintf("%s/photo/keyword/%s", jpHomeURL, reg)
	b, err := scraper.FetchHTML(URL)
	if err != nil {
		return nil, jpError("scraping search URL", reg, URL, err)
	}

	s := scraper.NewScraper(b)
	defer s.Close()

	pageLinks := []string{}
	thumbnails := []string{}
	for i := 0; i < q.Photos; i++ {
		pageLink, err := s.ScrapeLinks("a", "result__photoLink", 1)
		if err != nil {
			if len(pageLinks) > 0 {
				break
			}
			return nil, jpError("scraping aircraft pagelinks", reg, URL, err)
		}

		thumbnail, err := s.ScrapeLinks("img", "result__photo", 1)
		if err != nil {
			if len(thumbnails) > 0 {
				break
			}
			return nil, jpError("scraping aircraft thumbnails", reg, URL, err)
		}
		pageLinks = append(pageLinks, pageLink[0])
		thumbnails = append(thumbnails, thumbnail[0])
	}

	imgs := len(pageLinks)

	images := make([]ImageAttributes, imgs)

	for i, link := range pageLinks {
		photoURL := fmt.Sprintf("%s%s", jpHomeURL, link)
		images[i].Link = photoURL
		images[i].Thumbnail = "https:" + thumbnails[i]

		b, err := scraper.FetchHTML(photoURL)
		if err != nil {
			return nil, jpError("fetching HTML page", reg, URL, err)
		}

		s := scraper.NewScraper(b)
		defer s.Close()

		// photo links
		photoLinkArr, err := s.ScrapeLinks("img", "large-photo__img", 1)
		if err != nil {
			return nil, jpError("scraping photo links", reg, URL, err)
		}
		images[i].Image = photoLinkArr[0]

		// registration + dates
		res, err := s.ScrapeText("h4", "headerText4 color-shark", 3)
		if err != nil {
			return nil, jpError("scraping registrating text", reg, URL, err)
		}
		images[i].DateTaken = res[1]
		images[i].DateUploaded = res[2]

		// aircraft
		s.Advance("h2", "header-reset", 1)
		res, err = s.ScrapeText("a", "link", 3)
		if err != nil {
			return nil, jpError("scraping aircraft text", reg, URL, err)
		}
		images[i].Aircraft = res[0]
		images[i].Airline = res[1]
		images[i].Serial = strings.TrimSpace(res[2])

		// location
		s.Advance("h5", "header-reset", 1)
		location, err := s.ScrapeText("a", "link", 1)
		if err != nil {
			return nil, jpError("scraping location text", reg, URL, err)
		}
		images[i].Location = location[0]

		// photographer
		photographer, err := s.ScrapeText("h6", "header-reset", 1)
		if err != nil {
			return nil, jpError("scraping photographer text", reg, URL, err)
		}
		images[i].Photographer = photographer[0]
	}

	result := &JetPhotosResult{
		Images: images,
		Reg:    strings.ToUpper(reg),
	}

	return result, nil
}

func jpError(msg, reg, url string, err error) error {
	return fmt.Errorf("Error %s for %s at %s: %v", msg, reg, url, err)
}
