package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mugahq/muga/api/models"
)

// Client is a thin wrapper around the Muga API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates an unauthenticated API client for the given base URL.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewAuthenticatedClient creates an API client that sends a Bearer token.
func NewAuthenticatedClient(baseURL, token string) *Client {
	c := NewClient(baseURL)
	c.token = token
	return c
}

// APIError represents a structured error from the Muga API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (c *Client) newRequest(method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	return resp, nil
}

func (c *Client) post(path string, body io.Reader) (*http.Response, error) {
	req, err := c.newRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

func (c *Client) get(path string) (*http.Response, error) {
	req, err := c.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return c.do(req)
}

// checkResponse returns an APIError if the response status is not in the 2xx range.
func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	var errResp models.ErrorResponse
	if err := decodeJSON(resp.Body, &errResp); err != nil {
		return &APIError{
			Code:    http.StatusText(resp.StatusCode),
			Message: fmt.Sprintf("unexpected status %d", resp.StatusCode),
		}
	}
	return &APIError{
		Code:    string(errResp.Error.Code),
		Message: errResp.Error.Message,
	}
}

func decodeJSON(r io.Reader, v any) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
}
