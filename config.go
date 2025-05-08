package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Config holds all configuration for PhantomDNS
type Config struct {
	// DNS Server settings
	DNSPort     int      `json:"dns_port"`
	DNSListen   string   `json:"dns_listen"`
	Nameservers []string `json:"nameservers"`

	// Blessnet settings
	BlessnetWorkerURL string `json:"blessnet_worker_url"`
	BlessnetAPIKey    string `json:"blessnet_api_key"`
	BlessnetAPISecret string `json:"blessnet_api_secret"`

	// API configuration
	API struct {
		BaseURL string `json:"base_url"`
		Version string `json:"version"`
	} `json:"api"`

	// Auth configuration
	Auth struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Token    string `json:"token"`
	} `json:"auth"`

	// Deployment configuration
	Deployment struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	} `json:"deployment"`

	// Worker configuration
	Worker struct {
		Count      int               `json:"count"`
		Regions    []string          `json:"regions"`
		Attributes map[string]string `json:"attributes"`
	} `json:"worker"`

	// DNS blocking settings
	BlockedDomains []string `json:"blocked_domains"`
	ProxyDomains   []string `json:"proxy_domains"`

	// Proxy settings
	ProxyMode string `json:"proxy_mode"`
}

// LoadConfig loads the configuration from file
func LoadConfig(path string) (*Config, error) {
	configFile, err := os.Open(path)
	if err != nil {
		// If file doesn't exist, create with default settings
		if os.IsNotExist(err) {
			return createDefaultConfig(path)
		}
		return nil, err
	}
	defer configFile.Close()

	// Read config file
	configData, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	// Parse JSON into Config struct
	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	// Apply defaults for any missing values
	applyConfigDefaults(&config)

	return &config, nil
}

// createDefaultConfig creates a default configuration file
func createDefaultConfig(path string) (*Config, error) {
	// Ensure directory exists
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}

	// Create default config
	config := &Config{
		DNSPort:     53,
		DNSListen:   "127.0.0.1",
		Nameservers: []string{"8.8.8.8", "1.1.1.1"},

		BlessnetWorkerURL: "https://apricot-emu-jacklin-qikeha7m.bls.dev",

		API: struct {
			BaseURL string `json:"base_url"`
			Version string `json:"version"`
		}{
			BaseURL: "https://api.bless.network",
			Version: "v1",
		},

		Deployment: struct {
			ID  string `json:"id"`
			URL string `json:"url"`
		}{
			ID:  "bafybeidztwbl5w4ih6c6rd4qzl35jmxbm33wm3zd2htjffzxesqikeha7m",
			URL: "https://apricot-emu-jacklin-qikeha7m.bls.dev",
		},

		Worker: struct {
			Count      int               `json:"count"`
			Regions    []string          `json:"regions"`
			Attributes map[string]string `json:"attributes"`
		}{
			Count:   3,
			Regions: []string{"us-east", "eu-west", "ap-east"},
			Attributes: map[string]string{
				"purpose": "dns-proxy",
			},
		},

		ProxyMode: "ephemeral",
	}

	// Write config to file
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, err
	}

	err = ioutil.WriteFile(path, configData, 0644)
	if err != nil {
		return nil, err
	}

	return config, nil
}

// applyConfigDefaults applies default values to missing configuration fields
func applyConfigDefaults(config *Config) {
	// Apply DNS defaults if not set
	if config.DNSPort == 0 {
		config.DNSPort = 53
	}
	if config.DNSListen == "" {
		config.DNSListen = "127.0.0.1"
	}
	if len(config.Nameservers) == 0 {
		config.Nameservers = []string{"8.8.8.8", "1.1.1.1"}
	}

	// Apply Blessnet defaults if not set
	if config.BlessnetWorkerURL == "" {
		config.BlessnetWorkerURL = "https://apricot-emu-jacklin-qikeha7m.bls.dev"
	}

	// Apply API defaults if not set
	if config.API.BaseURL == "" {
		config.API.BaseURL = "https://api.bless.network"
	}
	if config.API.Version == "" {
		config.API.Version = "v1"
	}

	// Apply worker defaults if not set
	if config.Worker.Count == 0 {
		config.Worker.Count = 3
	}
	if len(config.Worker.Regions) == 0 {
		config.Worker.Regions = []string{"us-east", "eu-west", "ap-east"}
	}
	if config.Worker.Attributes == nil {
		config.Worker.Attributes = map[string]string{
			"purpose": "dns-proxy",
		}
	}

	// Apply proxy mode default if not set
	if config.ProxyMode == "" {
		config.ProxyMode = "ephemeral"
	}
}

// SaveConfig saves the configuration to a file
func SaveConfig(config *Config) error {
	// Determine the location of the configuration file
	configPath := "config.json"

	// Use environment variable if set
	if envPath := os.Getenv("PHANTOMDNS_CONFIG"); envPath != "" {
		configPath = envPath
	}

	// Convert configuration to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Create directory (if it doesn't exist)
	dir := filepath.Dir(configPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	// Write to file
	return ioutil.WriteFile(configPath, data, 0644)
}
