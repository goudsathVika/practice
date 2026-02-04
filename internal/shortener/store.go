package shortener

import (
	"sort"
	"sync"
)

type DomainCount struct {
	Domain string `json:"domain"`
	Count  int    `json:"count"`
}

type Store struct {
	mu           sync.Mutex
	byURL        map[string]string
	byCode       map[string]string
	domainCounts map[string]int
	counter      uint64
}

func NewStore() *Store {
	return &Store{
		byURL:        make(map[string]string),
		byCode:       make(map[string]string),
		domainCounts: make(map[string]int),
	}
}

func (s *Store) GetOrCreate(url, domain string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.domainCounts[domain]++
	if code, ok := s.byURL[url]; ok {
		return code
	}

	s.counter++
	code := encodeBase62(s.counter)
	s.byURL[url] = code
	s.byCode[code] = url
	return code
}

func (s *Store) GetOriginal(code string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	url, ok := s.byCode[code]
	return url, ok
}

func (s *Store) TopDomains(limit int) []DomainCount {
	s.mu.Lock()
	defer s.mu.Unlock()

	counts := make([]DomainCount, 0, len(s.domainCounts))
	for domain, count := range s.domainCounts {
		counts = append(counts, DomainCount{Domain: domain, Count: count})
	}

	sort.Slice(counts, func(i, j int) bool {
		if counts[i].Count == counts[j].Count {
			return counts[i].Domain < counts[j].Domain
		}
		return counts[i].Count > counts[j].Count
	})

	if limit > 0 && len(counts) > limit {
		return counts[:limit]
	}
	return counts
}

const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func encodeBase62(value uint64) string {
	if value == 0 {
		return string(base62Alphabet[0])
	}

	encoded := make([]byte, 0)
	for value > 0 {
		remainder := value % 62
		encoded = append(encoded, base62Alphabet[remainder])
		value /= 62
	}

	for i, j := 0, len(encoded)-1; i < j; i, j = i+1, j-1 {
		encoded[i], encoded[j] = encoded[j], encoded[i]
	}

	return string(encoded)
}
