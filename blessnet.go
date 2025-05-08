package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

// AuthConfig stores authentication information
type AuthConfig struct {
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// BlessnetConfig structure contains all configuration settings needed for communication with Blessnet
type BlessnetConfig struct {
	// Blessnet API properties
	API struct {
		BaseURL string // API endpoint URL
		Version string // API version
	}

	// Blessnet Authentication
	Auth struct {
		Token     string // Authentication token
		ExpiresAt time.Time
	}

	// Deployment Information
	Deployment struct {
		ID  string // Deployment ID (CID)
		URL string // Deployment URL (deployed function URL)
	}

	// Worker Configuration
	Worker struct {
		Count      int      // Number of workers to use
		Regions    []string // Regions where workers will operate
		Attributes map[string]string
	}
}

// BlessnetClient handles all communication with Blessnet
type BlessnetClient struct {
	Config     *Config
	WorkerURL  string
	Regions    []string
	ActiveNode string
	client     *http.Client
	mutex      sync.RWMutex
	auth       *AuthConfig
}

// NewBlessnetClient creates a new Blessnet client
func NewBlessnetClient(config *Config) (*BlessnetClient, error) {
	client := &BlessnetClient{
		Config:  config,
		Regions: []string{"us-east", "eu-west", "ap-east"},
		client:  &http.Client{Timeout: 30 * time.Second},
		auth:    &AuthConfig{},
	}

	// Get the worker URL from config or use the default from bls.toml
	client.WorkerURL = config.BlessnetWorkerURL
	if client.WorkerURL == "" {
		client.WorkerURL = "https://apricot-emu-jacklin-qikeha7m.bls.dev"
	}

	log.Printf("Initializing Blessnet client with worker URL: %s", client.WorkerURL)

	// Test the connection to the worker
	err := client.TestConnection()
	if err != nil {
		log.Printf("Warning: Initial connection to Blessnet worker failed: %v", err)
		log.Printf("Will try alternative methods or regions when needed")
	} else {
		log.Printf("Successfully connected to Blessnet worker")
	}

	return client, nil
}

// Authenticate with the Blessnet API
func (b *BlessnetClient) Authenticate() error {
	// Check if token is still valid
	if b.auth.Token != "" && b.auth.ExpiresAt.After(time.Now()) {
		log.Printf("Using existing token (expires in %v)", time.Until(b.auth.ExpiresAt))
		return nil
	}

	log.Printf("Authenticating with Blessnet API...")

	// In a real implementation, we would call the Blessnet API to get a token
	// For this demo, we'll just generate a fake token
	b.mutex.Lock()
	b.auth.Token = "simulated_token_" + time.Now().Format(time.RFC3339)
	b.auth.ExpiresAt = time.Now().Add(24 * time.Hour)
	b.mutex.Unlock()

	log.Printf("Authentication successful, token expires in %v", time.Until(b.auth.ExpiresAt))
	return nil
}

// RefreshAuth forces a refresh of the authentication token
func (b *BlessnetClient) RefreshAuth() error {
	b.mutex.Lock()
	b.auth.Token = ""
	b.mutex.Unlock()
	return b.Authenticate()
}

// TestConnection checks if the worker URL is accessible
func (b *BlessnetClient) TestConnection() error {
	_, err := fetchFromWorker("https://example.com")
	return err
}

// DeployWorker handles the deployment of the worker code
func (b *BlessnetClient) DeployWorker() error {
	log.Println("Deploying Blessnet worker...")

	// Check if blessnet CLI is installed
	_, err := exec.LookPath("blessnet")
	if err != nil {
		return fmt.Errorf("blessnet CLI not found. Please install it first: %v", err)
	}

	// Execute blessnet deploy command
	cmd := exec.Command("blessnet", "deploy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to deploy worker: %v", err)
	}

	log.Println("Worker deployed successfully")
	return nil
}

// RefreshNodes refreshes the node list to ensure we have active nodes
func (b *BlessnetClient) RefreshNodes() error {
	log.Println("Refreshing Blessnet nodes...")

	// This is a placeholder for actual node refresh logic
	// In a real implementation, you might query a Blessnet API
	// to get an updated list of available nodes

	return nil
}

// SendProxyRequest enables proxy functionality for a blocked domain
func (b *BlessnetClient) SendProxyRequest(targetURL string) ([]byte, error) {
	// Use the fetchFromWorker function for proxy requests
	return fetchFromWorker(targetURL)
}

// ListDeployments gets a list of all current deployments
func (b *BlessnetClient) ListDeployments() ([]string, error) {
	// This is a placeholder for a real API call
	// In a real implementation, you would call the Blessnet API

	log.Println("Listing deployments...")

	// Return a simulated list
	return []string{"apricot-emu-jacklin-qikeha7m.bls.dev"}, nil
}

// CreateWorkerTemplate returns a TypeScript template for creating a new worker
func (b *BlessnetClient) CreateWorkerTemplate() string {
	return `import { main } from "@blockless/sdk-ts/dist/lib/entry"; // Import directly from submodule

// Define a type for environment variables
interface EnvVars {
  TARGET?: string;
}

main(async () => {
  // Get environment variables and check if TARGET is empty or undefined
  const env: EnvVars = process.env as any;
  const targetUrl = env.TARGET || "";
  
  // Show info if no target is specified
  if (!targetUrl || targetUrl.trim() === "") {
    console.log("PhantomDNS Worker is active - Displaying welcome info");
    
    // Simple plain text response
    const welcomeInfo = 
      "PhantomDNS Worker is active\n" +
      "Worker ID: " + Math.random().toString(36).substring(2, 8) + "\n" +
      "Time: " + new Date().toISOString() + "\n\n" +
      "To use: Add TARGET parameter with the URL to access";

    return new Response(welcomeInfo, {
      status: 200,
      headers: {
        "Content-Type": "text/plain",
        "Cache-Control": "no-store, no-cache",
        "X-Proxy-By": "PhantomDNS"
      }
    });
  }
  
  console.log("PhantomDNS Worker active - Processing request for: " + targetUrl);

  try {
    console.log("Establishing connection: " + targetUrl);
    
    // Send request to external API using fetch
    // NOTE: Make sure this URL is in the permissions list in the bls.toml file
    const response = await fetch(targetUrl, { 
      method: 'GET', 
      headers: { 
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
        "Accept-Language": "en-US,en;q=0.9"
      }
    });

    console.log("Connection status: " + response.status);

    if (!response.ok) {
      // Return simple error text
      const errorInfo = 
        "ERROR: Failed to connect to target\n" +
        "Status: " + response.status + "\n" +
        "Target: " + targetUrl;
      
      return new Response(errorInfo, {
        status: 502,
        headers: {
          "Content-Type": "text/plain",
          "X-Proxy-By": "PhantomDNS"
        }
      });
    }

    // Get response as text
    const text = await response.text();
    console.log("Response received, size: " + text.length + " bytes");
    
    // Return simple success text with content sample
    const resultInfo = 
      "SUCCESS: Connected to target\n" +
      "Status: " + response.status + "\n" +
      "Target: " + targetUrl + "\n" +
      "Content Size: " + text.length + " bytes\n\n" +
      "Content Sample:\n" +
      "------------\n" +
      text.substring(0, 300) + (text.length > 300 ? "..." : "");
    
    return new Response(resultInfo, {
      status: 200,
      headers: {
        "Content-Type": "text/plain",
        "X-Proxy-By": "PhantomDNS"
      }
    });
  } catch (error) {
    console.error("Error: " + error.message);
    
    // Return simple error text
    const errorInfo = 
      "ERROR: Exception occurred\n" +
      "Message: " + error.message;
    
    return new Response(errorInfo, {
      status: 500,
      headers: {
        "Content-Type": "text/plain",
        "X-Proxy-By": "PhantomDNS"
      }
    });
  }
});`
}

// FetchPage retrieves content from a URL using the Blessnet worker
func (b *BlessnetClient) FetchPage(targetURL string) ([]byte, error) {
	return fetchFromWorker(targetURL)
}

// Using environment variables for API keys is more secure than hardcoding
func getEnvVar(key string) string {
	value := os.Getenv(key)
	if value == "" {
		// Log warning but don't expose in production logs
		log.Printf("Warning: Environment variable %s not set", key)
	}
	return value
}
