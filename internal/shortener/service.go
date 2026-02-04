package shortener

import (
	"fmt"
	"net/url"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) Shorten(rawURL string) (string, error) {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid url")
	}

	code := s.store.GetOrCreate(rawURL, parsed.Host)
	return code, nil
}

func (s *Service) Resolve(code string) (string, bool) {
	return s.store.GetOriginal(code)
}

func (s *Service) Metrics(limit int) []DomainCount {
	return s.store.TopDomains(limit)
}
