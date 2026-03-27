package api

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/mugahq/muga/api/models"
)

// ListProjects returns all projects the authenticated user has access to.
func (c *Client) ListProjects() ([]models.Project, error) {
	resp, err := c.get("/v1/projects")
	if err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	var result struct {
		Data       []models.Project `json:"data"`
		Pagination models.Pagination `json:"pagination"`
	}
	if err := decodeJSON(resp.Body, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

// CreateProject creates a new project.
func (c *Client) CreateProject(req models.CreateProjectRequest) (*models.Project, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("encoding create project request: %w", err)
	}

	resp, err := c.post("/v1/projects", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if err := checkResponse(resp); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}

	var project models.Project
	if err := decodeJSON(resp.Body, &project); err != nil {
		return nil, err
	}

	return &project, nil
}
