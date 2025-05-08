package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// BlessnetNodeAPI represents a specific Blessnet node API client
type BlessnetNodeAPI struct {
	BaseURL    string
	APIVersion string
	client     *http.Client
}

// NewBlessnetNodeAPI creates a new API client for a specific node
func NewBlessnetNodeAPI(baseURL string) *BlessnetNodeAPI {
	return &BlessnetNodeAPI{
		BaseURL:    baseURL,
		APIVersion: "v1",
		client:     &http.Client{Timeout: 30 * time.Second},
	}
}

// AuthResponse represents the response from authentication endpoint
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Auth authenticates with the Blessnet API
func (b *BlessnetNodeAPI) Auth(apiKey string, apiSecret string) (*AuthResponse, error) {
	// Create authentication request
	reqData := map[string]string{
		"api_key":    apiKey,
		"api_secret": apiSecret,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, fmt.Errorf("error marshaling auth request: %v", err)
	}

	// Send request to authentication endpoint
	resp, err := b.client.Post(
		fmt.Sprintf("%s/%s/auth", b.BaseURL, b.APIVersion),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("error sending auth request: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("auth request failed: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var authResp AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		return nil, fmt.Errorf("error parsing auth response: %v", err)
	}

	// Calculate expiration time
	authResp.ExpiresAt = time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	return &authResp, nil
}

// GetNodes retrieves available nodes
func (b *BlessnetNodeAPI) GetNodes() ([]map[string]interface{}, error) {
	// Add authentication
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf("%s/%s/nodes", b.BaseURL, b.APIVersion),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating nodes request: %v", err)
	}

	// Send request
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting nodes: %v", err)
	}
	defer resp.Body.Close()

	// Check response
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("get nodes request failed: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var nodesResp struct {
		Data []map[string]interface{} `json:"data"`
	}
	err = json.NewDecoder(resp.Body).Decode(&nodesResp)
	if err != nil {
		return nil, fmt.Errorf("error parsing nodes response: %v", err)
	}

	return nodesResp.Data, nil
}

// FetchNodeStatus checks the status of a specific Blessnet node
func (api *BlessnetNodeAPI) FetchNodeStatus(nodeID string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/nodes/%s", api.BaseURL, nodeID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create node status request: %v", err)
	}

	if api.APIVersion != "" {
		req.Header.Set("Authorization", "Bearer "+api.APIVersion)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send node status request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("node status query failed. Status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse node status response: %v", err)
	}

	return result, nil
}

// ListAvailableNodes lists all available Blessnet nodes
func (api *BlessnetNodeAPI) ListAvailableNodes() ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/nodes", api.BaseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create node list request: %v", err)
	}

	if api.APIVersion != "" {
		req.Header.Set("Authorization", "Bearer "+api.APIVersion)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send node list request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("node list query failed. Status code: %d", resp.StatusCode)
	}

	var result []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse node list response: %v", err)
	}

	return result, nil
}

// DeployFunction deploys a function to specific Blessnet nodes
func (api *BlessnetNodeAPI) DeployFunction(wasmBytes []byte, deployOptions map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/functions", api.BaseURL)

	// Create request body
	requestBody, err := json.Marshal(map[string]interface{}{
		"function": wasmBytes,
		"options":  deployOptions,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON for deploy request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create deploy request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if api.APIVersion != "" {
		req.Header.Set("Authorization", "Bearer "+api.APIVersion)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send deploy request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("deploy request failed. Status code: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse deploy response: %v", err)
	}

	return result, nil
}

// InvokeFunction calls a deployed function with specific parameters
func (api *BlessnetNodeAPI) InvokeFunction(functionID string, params map[string]interface{}) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/functions/%s/invoke", api.BaseURL, functionID)

	// Create request body
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON for invoke request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create invoke request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if api.APIVersion != "" {
		req.Header.Set("Authorization", "Bearer "+api.APIVersion)
	}

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send invoke request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Read response content (for error message)
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invoke request failed. Status code: %d, Response: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse invoke response: %v", err)
	}

	return result, nil
}

// DetectNodeEndpoints detects Blessnet nodes in the local network or known hosts
func (api *BlessnetNodeAPI) DetectNodeEndpoints() ([]string, error) {
	log.Println("Searching for Blessnet nodes in local network or known hosts...")

	// NOTE: In a real application, node discovery would be done here
	// For now, let's return a fixed list
	knownEndpoints := []string{
		"https://apricot-emu-jacklin-qikeha7m.bls.dev",
		"https://api.bless.network",
	}

	// Try to get node information from Blessnet CLI
	cmd := exec.Command("blessnet", "list", "nodes")
	output, err := cmd.Output()
	if err == nil {
		// Parse CLI output and add endpoints
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "http") {
				parts := strings.Fields(line)
				for _, part := range parts {
					if strings.HasPrefix(part, "http") {
						knownEndpoints = append(knownEndpoints, part)
					}
				}
			}
		}
	}

	return knownEndpoints, nil
}

// ConnectivityCheck performs a connection check to a Blessnet node
func (api *BlessnetNodeAPI) ConnectivityCheck(endpoint string) (bool, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create connection check request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("connection check failed: %v", err)
	}
	defer resp.Body.Close()

	return resp.StatusCode < 500, nil // 4xx or 2xx status codes indicate that the server is at least running
}
