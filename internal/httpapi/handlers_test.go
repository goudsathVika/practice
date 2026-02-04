package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"practice/internal/shortener"
)

type shortenResponsePayload struct {
	ShortURL string `json:"short_url"`
}

type metricsResponsePayload struct {
	Domains []shortener.DomainCount `json:"domains"`
}

func newTestServer() *httptest.Server {
	store := shortener.NewStore()
	service := shortener.NewService(store)
	handler := NewHandler(service, "")
	mux := http.NewServeMux()
	handler.Register(mux)
	return httptest.NewServer(mux)
}

func TestShortenReturnsSameURL(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	payload := []byte(`{"url":"https://example.com/docs"}`)
	short1 := shortenURL(t, server.URL, payload)
	short2 := shortenURL(t, server.URL, payload)

	if short1 != short2 {
		t.Fatalf("expected same short url, got %q and %q", short1, short2)
	}
}

func TestRedirectToOriginal(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	shortURL := shortenURL(t, server.URL, []byte(`{"url":"https://example.com/hello"}`))

	req, err := http.NewRequest(http.MethodGet, shortURL, nil)
	if err != nil {
		t.Fatalf("failed to build request: %v", err)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		t.Fatalf("expected status 302, got %d", resp.StatusCode)
	}

	location := resp.Header.Get("Location")
	if location != "https://example.com/hello" {
		t.Fatalf("expected redirect location to original url, got %q", location)
	}
}

func TestMetricsTopDomains(t *testing.T) {
	server := newTestServer()
	defer server.Close()

	shortenURL(t, server.URL, []byte(`{"url":"https://udemy.com/course/1"}`))
	shortenURL(t, server.URL, []byte(`{"url":"https://udemy.com/course/2"}`))
	shortenURL(t, server.URL, []byte(`{"url":"https://udemy.com/course/3"}`))
	shortenURL(t, server.URL, []byte(`{"url":"https://youtube.com/watch?v=1"}`))
	shortenURL(t, server.URL, []byte(`{"url":"https://youtube.com/watch?v=2"}`))
	shortenURL(t, server.URL, []byte(`{"url":"https://wikipedia.org/wiki/1"}`))

	resp, err := http.Get(server.URL + "/metrics")
	if err != nil {
		t.Fatalf("metrics request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var payload metricsResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(payload.Domains) != 3 {
		t.Fatalf("expected 3 domains, got %d", len(payload.Domains))
	}

	if payload.Domains[0].Domain != "udemy.com" || payload.Domains[0].Count != 3 {
		t.Fatalf("expected udemy.com with count 3, got %+v", payload.Domains[0])
	}

	if payload.Domains[1].Domain != "youtube.com" || payload.Domains[1].Count != 2 {
		t.Fatalf("expected youtube.com with count 2, got %+v", payload.Domains[1])
	}
}

func shortenURL(t *testing.T, baseURL string, payload []byte) string {
	t.Helper()

	resp, err := http.Post(baseURL+"/shorten", "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("shorten request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var response shortenResponsePayload
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.ShortURL == "" {
		t.Fatal("short url was empty")
	}

	return response.ShortURL
}
