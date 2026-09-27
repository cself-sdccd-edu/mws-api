package qas

import (
	"context"
	"github.com/cself-sdccd-edu/mws-api/internal/config"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newTestClient(statusCode int, body string) *HTTPClient {
	httpClient := &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: statusCode,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(body)),
				Request:    r,
			}, nil
		}),
	}
	return NewHTTPClient(
		httpClient,
		"https://example.invalid",
		"/query/{QUERY_NAME}?",
		[]config.QueryParam{
			{
				Name:  "prompt_fieldvalue",
				Value: "{TERM}",
			},
		},
		"test-user",
		"test-password",
	)
}

func TestQueryRejectsEmptyResponse(t *testing.T) {
	client := newTestClient(http.StatusOK, "")

	_, err := client.Query(context.Background(), "TEST_QUERY", "2267")
	if err == nil {
		t.Fatal("expected empty QAS response to return an error")
	}
}

func TestQueryRejectsWhitespaceResponse(t *testing.T) {
	client := newTestClient(http.StatusOK, " \n\t ")

	_, err := client.Query(context.Background(), "TEST_QUERY", "2267")
	if err == nil {
		t.Fatal("expected whitespace QAS response to return an error")
	}
}

func TestQueryRejectsInvalidJSON(t *testing.T) {
	client := newTestClient(http.StatusOK, "<html>Login required</html>")

	_, err := client.Query(context.Background(), "TEST_QUERY", "2267")
	if err == nil {
		t.Fatal("expected invalid JSON to return an error")
	}
}

func TestQueryAcceptsValidJSON(t *testing.T) {
	client := newTestClient(http.StatusOK, `{"records":[{"term":"2267"}]}`)

	data, err := client.Query(context.Background(), "TEST_QUERY", "2267")
	if err != nil {
		t.Fatalf("expected valid JSON, got error: %v", err)
	}

	if string(data) != `{"records":[{"term":"2267"}]}` {
		t.Fatalf("unexpected response: %s", data)
	}
}

func TestQueryRejectsUnsuccessfulStatus(t *testing.T) {
	client := newTestClient(http.StatusBadGateway, `{"error":"upstream unavailable"}`)

	_, err := client.Query(context.Background(), "TEST_QUERY", "2267")
	if err == nil {
		t.Fatal("expected unsuccessful QAS status to return an error")
	}
}
