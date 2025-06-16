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
	tokens    []html.Token
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
	data := make([]string, len(tokens))
	for i := 0; i < len(data); i++ {
		data[i] = tokens[i].Data
	}
	return data, nil
}

func (s *Scraper) Advance(startTag, class string, count int) error {
	_, err := s.scrapeNextTokens(startTag, class, count, ADVANCE, html.StartTagToken)
	return err
}

func (s *Scraper) TryScrapeText() (string, bool) {
	tt := s.tokenizer.Next()
	t := s.tokenizer.Token()

	for strings.TrimSpace(t.Data) == "" {
		tt = s.tokenizer.Next()
		t = s.tokenizer.Token()
	}

	if tt != html.TextToken {
		s.tokens = append(s.tokens, t)
		return "", false
	}

	return t.Data, true
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
	var resultTokens []html.Token
	atLeastOne := false

	for count > 0 {
		token, err := s.nextToken(html.StartTagToken, startTag, class)

		if err != nil {
			if atLeastOne {
				break
			}
			return nil, err
		}

		if tt == html.TextToken {
			s.tokenizer.Next()
			token = s.tokenizer.Token()
		}

		if action == SCRAPE {
			resultTokens = append(resultTokens, token)
			atLeastOne = true
		}

		count--
	}

	return resultTokens, nil
}

func (s *Scraper) nextToken(tt html.TokenType, data, class string) (html.Token, error) {
	for i, t := range s.tokens {
		if t.Type == tt && t.Data == data && tokenHasClass(&t, class) {
			s.tokens = s.tokens[i+1:]
			return t, nil
		}
	}

	var nilToken html.Token
	for {
		tokenType := s.tokenizer.Next()
		if tokenType == html.ErrorToken {
			if s.tokenizer.Err() == io.EOF {
				return nilToken, s.Errorf("tag '%s' with class '%s' not found", data, class)
			}
			return nilToken, s.Errorf("Error tokenizing html: %v", s.tokenizer.Err())
		}
		t := s.tokenizer.Token()
		s.tokens = append(s.tokens, t)

		if tokenType == tt && t.Data == data && tokenHasClass(&t, class) {
			s.tokens = s.tokens[len(s.tokens):]
			return t, nil
		}
	}
}

func (s *Scraper) Errorf(format string, a ...any) error {
	return fmt.Errorf("Scraper Error: %s", fmt.Sprintf(format, a...))
}

func tokenHasClass(t *html.Token, class string) bool {
	if class == "" {
		return true
	}

	attrs := t.Attr

	for _, attr := range attrs {
		if attr.Key == "class" && attr.Val == class {
			return true
		}
	}
	return false
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
