package scraper

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

type Scraper struct {
	body      io.ReadCloser
	tokenizer *html.Tokenizer
}

type ActionType uint32

const (
	FETCH ActionType = iota
	ADVANCE
)

func NewScraper(body io.ReadCloser) *Scraper {
	return &Scraper{body: body}
}

func (s *Scraper) Close() {
	s.body.Close()
}

func (s *Scraper) FetchLinks(startTag, class string, quantity int) ([]string, error) {
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

func (s *Scraper) FetchText(startTag, class string, quantity int) ([]string, error) {
	tokens, err := s.fetchNextTokens(startTag, class, quantity, FETCH, html.TextToken)
	if err != nil {
		return nil, err
	}
	if len(tokens) != quantity {
		return nil, fmt.Errorf("text not found")
	}
	data := make([]string, len(tokens))
	for i := 0; i < quantity; i++ {
		data[i] = tokens[i].Data
	}
	return data, nil
}

func (s *Scraper) Advance(startTag, class string, quantity int) error {
	_, err := s.fetchNextTokens(startTag, class, quantity, ADVANCE, html.StartTagToken)
	return err
}

func (s *Scraper) fetchNextTokens(
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
				return nil, fmt.Errorf("query not found")
			}
			return nil, fmt.Errorf("error tokenizing html %v", s.tokenizer.Err())
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

func FetchHTML(URL string) (io.ReadCloser, error) {
	resp, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response error code: %v", resp.StatusCode)
	}
	ctype := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ctype, "text/html") {
		return nil, fmt.Errorf("content not type text/html")
	}
	return resp.Body, nil
}
