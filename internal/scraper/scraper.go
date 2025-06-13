package scraper

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

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
	donejp := make(chan jetPhotosRes)
	donefr := make(chan flightRadarRes)

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

type scraper struct {
	body      io.ReadCloser
	tokenizer *html.Tokenizer
}

type ActionType uint32

const (
	FETCH ActionType = iota
	ADVANCE
)

func newScraper(body io.ReadCloser) *scraper {
	return &scraper{body: body}
}

func (s *scraper) close() {
	s.body.Close()
}

func (s *scraper) fetchLinks(startTag, class string, quantity int) ([]string, error) {
	tokens, err := s.fetchNextTokens(startTag, class, quantity, FETCH, html.StartTagToken)
	if err != nil {
		return nil, err
	}
	links := make([]string, len(tokens))
	for i, tk := range tokens {
		for _, attr := range tk.Attr {
			if attr.Key == "href" || attr.Key == "src" || attr.Key == "srcset" {
				links[i] = attr.Val
				break
			}
		}
	}
	return links, nil
}

func (s *scraper) fetchText(startTag, class string, quantity int) ([]string, error) {
	tokens, err := s.fetchNextTokens(startTag, class, quantity, FETCH, html.TextToken)
	if err != nil {
		return nil, err
	}
	if len(tokens) != quantity {
		return nil, s.Errorf("text not found with start tag %s, class %s, wanted %d, got %d",
			startTag, class, quantity, len(tokens))
	}
	data := make([]string, len(tokens))
	for i := 0; i < quantity; i++ {
		data[i] = tokens[i].Data
	}
	return data, nil
}

func (s *scraper) advance(startTag, class string, quantity int) error {
	_, err := s.fetchNextTokens(startTag, class, quantity, ADVANCE, html.StartTagToken)
	return err
}

func (s *scraper) fetchNextTokens(
	startTag, class string,
	quantity int,
	action ActionType,
	tt html.TokenType,
) ([]html.Token, error) {
	if s.tokenizer == nil {
		s.tokenizer = html.NewTokenizer(s.body)
	}
	var tokens []html.Token
	initQuant := quantity
	atLeastOne := false
	for {
		tokenType := s.tokenizer.Next()
		if tokenType == html.ErrorToken {
			if s.tokenizer.Err() == io.EOF {
				if atLeastOne {
					break
				}
				return nil, s.Errorf("unexpected EOF")
			}
			return nil, s.Errorf("error tokenizing html %s", s.tokenizer.Err().Error())
		}

		if tokenType != html.StartTagToken {
			continue
		}

		token := s.tokenizer.Token()
		if token.Data != startTag {
			continue
		}

		attr := token.Attr

		if class != "" {
			classPos := -1
			for i := 0; i < len(attr); i++ {
				if attr[i].Key == "class" {
					classPos = i
					break
				}
			}
			if classPos == -1 || attr[classPos].Val != class {
				continue
			}
		}

		if tt == html.TextToken {
			s.tokenizer.Next()
			token = s.tokenizer.Token()
		}

		if action == FETCH {
			tokens = append(tokens, token)
			atLeastOne = true
		}

		quantity--
		if quantity == 0 {
			break
		}
	}
	if initQuant < 0 {
		lastTokens := make([]html.Token, -initQuant)
		ttsize := len(tokens)
		for i := ttsize + initQuant; i < ttsize; i++ {
			lastTokens[i-ttsize-initQuant] = tokens[i]
		}
		return lastTokens, nil
	}
	return tokens, nil
}

func (s *scraper) Errorf(format string, a ...any) error {
	return fmt.Errorf("Scraper Error: %s", fmt.Sprintf(format, a...))
}

func fetchHTML(URL string) (io.ReadCloser, error) {
	tlsConfig := &tls.Config{
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		},
		MinVersion:       tls.VersionTLS12,
		MaxVersion:       tls.VersionTLS13,
		CurvePreferences: []tls.CurveID{tls.CurveP256, tls.X25519},
	}

	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %s", err.Error())
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	var resp *http.Response

	// retry 3 times, sometimes it returns 403
	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Error sending request: %s", err.Error())
		}

		if resp.StatusCode == http.StatusOK {
			break
		} else if i < 2 && resp.StatusCode == http.StatusForbidden {
			continue
		} else {
			return nil, fmt.Errorf("response error code: %v, URL: %s", resp.StatusCode, URL)
		}
	}

	ctype := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "text/html") {
		return nil, fmt.Errorf("content not type text/html")
	}

	return resp.Body, nil
}
