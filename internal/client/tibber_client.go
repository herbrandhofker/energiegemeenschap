package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// TibberClient provides a simple client for the Tibber GraphQL API
type TibberClient struct {
	APIToken  string
	APIURL    string
	UserAgent string
}

// GraphQLResponse represents a response from the Tibber GraphQL API
type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// NewClient creates a new Tibber client with the given API token
func NewClient(apiToken string) *TibberClient {
	return &TibberClient{
		APIToken:  apiToken,
		APIURL:    "https://api.tibber.com/v1-beta/gql",
		UserAgent: "TibberClient/1.0",
	}
}

// QueryAPI executes a GraphQL query against the Tibber API
func (c *TibberClient) QueryAPI(ctx context.Context, query string, variables map[string]interface{}) (*GraphQLResponse, error) {
	// Build request body
	reqBody := map[string]interface{}{"query": query}
	if len(variables) > 0 {
		reqBody["variables"] = variables
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Create and execute HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.APIURL, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIToken)
	req.Header.Set("User-Agent", c.UserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Process response
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// Parse the response
	var graphqlResp GraphQLResponse
	if err := json.Unmarshal(bodyBytes, &graphqlResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(graphqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphqlResp.Errors[0].Message)
	}

	return &graphqlResp, nil
}

// Helper function to safely extract string values
func GetString(data map[string]interface{}, key string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return ""
}

// Helper function to safely extract int values
func GetInt(data map[string]interface{}, key string) int {
	if val, ok := data[key].(float64); ok {
		return int(val)
	}
	return 0
}
