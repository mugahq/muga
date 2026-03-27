package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a thin wrapper around the Muga API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates an API client for the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIError represents a structured error from the Muga API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (c *Client) post(path string, body io.Reader) (*http.Response, error) {
	url := c.baseURL + path

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}

	return resp, nil
}

func decodeJSON(r io.Reader, v any) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}
