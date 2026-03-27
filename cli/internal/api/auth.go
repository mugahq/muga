package api

import (
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

// TokenResponse is the response from POST /auth/token.
type TokenResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	UserName     string `json:"user_name,omitempty"`
	UserEmail    string `json:"user_email,omitempty"`
	Error        string `json:"error,omitempty"`
}

// RequestDeviceCode starts the device authorization flow.
func (c *Client) RequestDeviceCode() (*DeviceCodeResponse, error) {
	resp, err := c.post("/auth/device", nil)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requesting device code: unexpected status %d", resp.StatusCode)
	}

	var result DeviceCodeResponse
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PollToken polls for a token using the device code.
func (c *Client) PollToken(deviceCode string) (*TokenResponse, error) {
	resp, err := c.post("/auth/token?device_code="+deviceCode, nil)
	if err != nil {
		return nil, fmt.Errorf("polling token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusForbidden {
		return nil, fmt.Errorf("polling token: unexpected status %d", resp.StatusCode)
	}

	var result TokenResponse
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
