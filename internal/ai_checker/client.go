package ai_checker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the AI service
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

// NewClient creates a new AI checker client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second, // AI analysis can take time
		},
	}
}

// ProposalCheckRequest is the request body for proposal analysis
type ProposalCheckRequest struct {
	Title      string `json:"title"`
	Objectives string `json:"objectives"`
}

// DetailedSummary contains the detailed analysis summary
type DetailedSummary struct {
	ProblemStatement string `json:"problem_statement"`
	ProposedSolution string `json:"proposed_solution"`
}

// RiskAssessment contains risk analysis data
type RiskAssessment struct {
	FeasibilityScore float64  `json:"feasibility_score"`
	TechnicalRisks   []string `json:"technical_risks"`
	Recommendations  []string `json:"recommendations"`
}

// MethodologyAnalysis contains methodology analysis
type MethodologyAnalysis struct {
	Strengths   []string `json:"strengths"`
	Weaknesses  []string `json:"weaknesses"`
	Suggestions []string `json:"suggestions"`
}

// SimilarityWarning represents a similar project warning
type SimilarityWarning struct {
	ProjectID       int      `json:"project_id"`
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	SimilarityScore float64  `json:"similarity_score"`
	CommonThemes    []string `json:"common_themes"`
	Explanation     string   `json:"explanation"`
	Suggestion      string   `json:"suggestion"`
}

// ProposalCheckResponse is the response from proposal analysis
type ProposalCheckResponse struct {
	Summary            string              `json:"summary"`
	DetailedSummary    DetailedSummary     `json:"detailed_summary"`
	Keywords           []string            `json:"keywords"`
	StructureHints     []string            `json:"structure_hints"`
	SimilarityWarnings []SimilarityWarning `json:"similarity_warnings"`
	RiskAssessment     RiskAssessment      `json:"risk_assessment"`
	MethodologyAnalysis MethodologyAnalysis `json:"methodology_analysis"`
	ConfidenceScores   map[string]float64  `json:"confidence_scores"`
}

// ProjectSyncItem represents a project to sync with the AI service
type ProjectSyncItem struct {
	ID      uint   `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

// SyncResponse is the response from project sync
type SyncResponse struct {
	Status          string `json:"status"`
	ProjectsIndexed int    `json:"projects_indexed"`
	Message         string `json:"message"`
}

// CheckProposal analyzes a proposal using the AI service
func (c *Client) CheckProposal(title, objectives string) (*ProposalCheckResponse, error) {
	url := fmt.Sprintf("%s/api/v1/predict/proposal-check", c.baseURL)

	payload := ProposalCheckRequest{
		Title:      title,
		Objectives: objectives,
	}

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ProposalCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// SyncProjects syncs approved projects to the AI service for similarity detection
func (c *Client) SyncProjects(projects []ProjectSyncItem) (*SyncResponse, error) {
	url := fmt.Sprintf("%s/api/v1/internal/sync-projects", c.baseURL)

	jsonBody, err := json.Marshal(projects)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AI service sync request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service sync returned status %d: %s", resp.StatusCode, string(body))
	}

	var result SyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode sync response: %w", err)
	}

	return &result, nil
}

// HealthCheck checks if the AI service is available
func (c *Client) HealthCheck() error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("AI service health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI service unhealthy, status: %d", resp.StatusCode)
	}

	return nil
}
