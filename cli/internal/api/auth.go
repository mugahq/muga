package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mugahq/muga/api/models"
)

// RequestDeviceCode starts the device authorization flow.
func (c *Client) RequestDeviceCode() (*models.DeviceFlowResponse, error) {
	resp, err := c.post("/auth/device", nil)
	if err != nil {
		return nil, fmt.Errorf("requesting device code: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requesting device code: unexpected status %d", resp.StatusCode)
	}

	var result models.DeviceFlowResponse
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// PollToken polls for a token using the device code.
func (c *Client) PollToken(deviceCode string) (*models.PollTokenResponse, error) {
	body, err := json.Marshal(models.PollTokenRequest{DeviceCode: deviceCode})
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

	var result models.PollTokenResponse
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
