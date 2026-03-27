package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// DeviceCodeResponse is the response from POST /auth/device.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// TokenUser is the nested user object inside PollTokenResponse.
type TokenUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Tier  string `json:"tier"`
}

// TokenResponse is the response from POST /auth/token.
type TokenResponse struct {
	Status       string     `json:"status"`
	AccessToken  string     `json:"access_token,omitempty"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	ExpiresIn    int        `json:"expires_in,omitempty"`
	User         *TokenUser `json:"user,omitempty"`
}

// RequestDeviceCode starts the device authorization flow.
func (c *Client) RequestDeviceCode() (*DeviceCodeResponse, error) {
	resp, err := c.post("/auth/device", nil)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requesting device code: unexpected status %d", resp.StatusCode)
	}

	var result DeviceCodeResponse
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// pollTokenRequest is the JSON body for POST /auth/token.
type pollTokenRequest struct {
	DeviceCode string `json:"device_code"`
}

// PollToken polls for a token using the device code.
func (c *Client) PollToken(deviceCode string) (*TokenResponse, error) {
	body, err := json.Marshal(pollTokenRequest{DeviceCode: deviceCode})
	if err != nil {
		return nil, fmt.Errorf("encoding poll request: %w", err)
	}

	resp, err := c.post("/auth/token", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("polling token: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("polling token: unexpected status %d", resp.StatusCode)
	}

	var result TokenResponse
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
