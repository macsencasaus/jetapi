package scraper

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

type Scraper struct {
	body      io.ReadCloser
	tokenizer *html.Tokenizer
}

type ActionType uint32

const (
	SCRAPE ActionType = iota
	ADVANCE
)

func NewScraper(body io.ReadCloser) *Scraper {
	return &Scraper{body: body}
}

func (s *Scraper) Close() {
	s.body.Close()
}

func (s *Scraper) ScrapeLinks(startTag, class string, count int) ([]string, error) {
	tokens, err := s.scrapeNextTokens(startTag, class, count, SCRAPE, html.StartTagToken)
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

func (s *Scraper) ScrapeText(startTag, class string, count int) ([]string, error) {
	tokens, err := s.scrapeNextTokens(startTag, class, count, SCRAPE, html.TextToken)
	if err != nil {
		return nil, err
	}
	if len(tokens) != count {
		return nil, s.Errorf("text not found with start tag %s, class %s, wanted %d, got %d",
			startTag, class, count, len(tokens))
	}
	data := make([]string, len(tokens))
	for i := 0; i < count; i++ {
		data[i] = tokens[i].Data
	}
	return data, nil
}

func (s *Scraper) Advance(startTag, class string, count int) error {
	_, err := s.scrapeNextTokens(startTag, class, count, ADVANCE, html.StartTagToken)
	return err
}

func (s *Scraper) scrapeNextTokens(
	startTag, class string,
	count int,
	action ActionType,
	tt html.TokenType,
) ([]html.Token, error) {
	if s.tokenizer == nil {
		s.tokenizer = html.NewTokenizer(s.body)
	}
	var tokens []html.Token
	atLeastOne := false

	for count > 0 {
		tokenType := s.tokenizer.Next()
		if tokenType == html.ErrorToken {
			if s.tokenizer.Err() == io.EOF {
				if atLeastOne {
					break
				}
				return nil, s.Errorf("unexpected EOF")
			}
			return nil, s.Errorf("Error tokenizing html: %v", s.tokenizer.Err())
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

		if action == SCRAPE {
			tokens = append(tokens, token)
			atLeastOne = true
		}

		count--
	}

	return tokens, nil
}

func (s *Scraper) Errorf(format string, a ...any) error {
	return fmt.Errorf("Scraper Error: %s", fmt.Sprintf(format, a...))
}

func FetchHTML(URL string) (io.ReadCloser, error) {
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
		return nil, fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	var resp *http.Response

	// retry 3 times, sometimes it returns 403
	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Error sending request: %v", err)
		}

		if resp.StatusCode == http.StatusOK {
			break
		} else if i < 2 && resp.StatusCode == http.StatusForbidden {
			continue
		} else {
			return nil, fmt.Errorf("response error code: %d, URL: %s", resp.StatusCode, URL)
		}
	}

	ctype := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "text/html") {
		return nil, fmt.Errorf("content not type text/html")
	}

	return resp.Body, nil
}
