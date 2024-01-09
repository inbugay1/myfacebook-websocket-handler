package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type APIClient struct {
	BaseURL string
	client  HTTPClient
}

func New(baseURL string, client HTTPClient) *APIClient {
	return &APIClient{
		BaseURL: baseURL,
		client:  client,
	}
}

func (c *APIClient) Request(ctx context.Context, method, url string, body io.Reader) (*http.Response, error) {
	url = c.BaseURL + url

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("apiclient could not create request: %w", err)
	}

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("apiclient could not send request: %w", err)
	}

	return response, nil
}

func (c *APIClient) Get(ctx context.Context, url string) (*http.Response, error) {
	response, err := c.Request(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("apiclient GET request failed: %w", err)
	}

	return response, nil
}

func (c *APIClient) Post(ctx context.Context, url string, data interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("apiclient could not marshal data: %w", err)
	}

	response, err := c.Request(ctx, http.MethodPost, url, bytes.NewBuffer(jsonData))

	if err != nil {
		return nil, fmt.Errorf("apiclient POST request failed: %w", err)
	}

	return response, nil
}
