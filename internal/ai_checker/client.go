package ai_checker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewClient(baseURL, apiKey string) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) Health(ctx context.Context) error {
	if c.baseURL == "" {
		return errors.New("AI service URL is not configured")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI service health check failed: %s", strings.TrimSpace(string(body)))
	}

	return nil
}

func (c *Client) CheckProposalText(ctx context.Context, payload ProposalCheckRequest) (map[string]interface{}, error) {
	if c.baseURL == "" {
		return nil, errors.New("AI service URL is not configured")
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/predict/proposal-check", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	applyHeaders(req, "application/json", c.apiKey)

	return c.doJSON(req)
}

func (c *Client) CheckProposalFile(ctx context.Context, filename string, fileContent []byte) (map[string]interface{}, error) {
	if c.baseURL == "" {
		return nil, errors.New("AI service URL is not configured")
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(fileContent); err != nil {
		return nil, err
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/predict/proposal-check-file", body)
	if err != nil {
		return nil, err
	}
	applyHeaders(req, writer.FormDataContentType(), c.apiKey)

	return c.doJSON(req)
}

func (c *Client) SyncProjects(ctx context.Context, projects []SyncProject) error {
	if c.baseURL == "" {
		return errors.New("AI service URL is not configured")
	}

	jsonBody, err := json.Marshal(projects)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/internal/sync-projects", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	applyHeaders(req, "application/json", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI service sync failed: %s", strings.TrimSpace(string(body)))
	}

	return nil
}

func (c *Client) doJSON(req *http.Request) (map[string]interface{}, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service error: %s", strings.TrimSpace(string(body)))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func applyHeaders(req *http.Request, contentType, apiKey string) {
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}
}
