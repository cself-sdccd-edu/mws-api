package qas

import (
	"context"
	"fmt"
	"github.com/cself-sdccd-edu/mws-api/internal/config"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type Client interface {
	Query(ctx context.Context, queryName string, term string) ([]byte, error)
}

type HTTPClient struct {
	client      *http.Client
	domain      string
	urlSuffix   string
	queryParams []config.QueryParam
	user        string
	password    string
}

func NewHTTPClient(client *http.Client, domain string, urlSuffix string, queryParams []config.QueryParam, user string, password string) *HTTPClient {
	return &HTTPClient{
		client:      client,
		domain:      domain,
		urlSuffix:   urlSuffix,
		queryParams: queryParams,
		user:        user,
		password:    password,
	}
}

func (c *HTTPClient) Query(ctx context.Context, queryName string, term string) ([]byte, error) {
	suffix := strings.ReplaceAll(c.urlSuffix, "{QUERY_NAME}", url.PathEscape(queryName))

	var query []string
	for _, param := range c.queryParams {
		value := strings.ReplaceAll(param.Value, "{TERM}", term)
		query = append(query, url.QueryEscape(param.Name)+"="+url.QueryEscape(value))
		//log.Default().Printf("query param: %s = %s", param.Name, value)
	}
	if len(query) == 0 {
		return nil, fmt.Errorf("no query params parsed")
	}

	requestURL := strings.TrimRight(c.domain, "/") + suffix + strings.Join(query, "&")

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create QAS request: %w", err)
	}
	// DEBUG: remote this output later
	log.Default().Printf("Request URL: %s", requestURL)

	request.SetBasicAuth(c.user, c.password)

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("execute QAS request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return nil, fmt.Errorf("QAS returned HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read QAS response: %w", err)
	}

	return data, nil
}
