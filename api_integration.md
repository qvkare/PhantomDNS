# Blessnet API Integration and API Key Acquisition Guide

This guide will help you integrate PhantomDNS with the Blessnet API and obtain API keys.

## 1. Creating a Blessnet Account

1. Create a Blessnet account if you don't have one: https://bless.network/signup
2. Verify your account and log in
3. Go to profile settings and set up your account

## 2. Configuring API Access

1. Go to the Blessnet Developer Portal
2. Find the "API Access" or "API Keys" section
3. Request or generate your API keys
4. Store your API keys securely

## 3. Setting Up Configuration

Update your `config.json` file with your Blessnet API information:

```json
{
  "dns_port": 5355,
  "dns_listen": "127.0.0.1",
  "nameservers": ["8.8.8.8", "1.1.1.1"],
  "blessnet_worker_url": "https://your-worker-url.bls.dev",
  "blessnet_api_key": "YOUR_API_KEY",
  "blessnet_api_secret": "YOUR_API_SECRET",
  "api": {
    "base_url": "https://api.bless.network",
    "version": "v1"
  }
}
```

## 4. API Integration with Blessnet CLI

To configure your API keys using the Blessnet CLI:

```bash
# Set API keys
blessnet config set api_key YOUR_API_KEY
blessnet config set api_secret YOUR_API_SECRET

# Verify API configuration
blessnet config get

# Test API authentication
blessnet auth status
```

## 5. bls.toml Configuration for External API Access

PhantomDNS uses a `bls.toml` file to specify permissions for external API access:

```toml
name = "phantomdns"
version = "1.0.0"
type = "text"
production_host = "apricot-emu-jacklin-qikeha7m.bls.dev"

[deployment]
permission = "public"
nodes = 3
permissions = [
  "https://api.bless.network/",
  "https://us-east.api.bless.network/",
  "https://eu-west.api.bless.network/",
  "https://ap-east.api.bless.network/",
  "https://node.bless.network/",
  "https://*",
  "http://*",
  "https://blockless.network/",
  "https://bls.dev/"
]

[build]
dir = "build"
entry = "debug.wasm"
command = "npm run build:debug"

[build_release]
dir = "build"
entry = "release.wasm"
command = "npm run build:release"
```

Important notes:
- Permissions are locked at deployment time and cannot be changed later
- You must include the full URLs with the protocol (https://)
- Only HTTPS is supported for secure endpoints
- The wildcard permissions (`https://*`, `http://*`) allow access to any domain, which is necessary for a proxy service

## 6. API Integration in PhantomDNS Code

PhantomDNS already includes Blessnet API integration through the `BlessnetClient` in `blessnet.go`:

```go
// Authenticate with the Blessnet API
func (b *BlessnetClient) Authenticate() error {
    // Check if token is still valid
    if b.auth.Token != "" && b.auth.ExpiresAt.After(time.Now()) {
        log.Printf("Using existing token (expires in %v)", time.Until(b.auth.ExpiresAt))
        return nil
    }

    log.Printf("Authenticating with Blessnet API...")

    // In a production implementation, we would call the Blessnet API to get a token
    // using the BlessnetClient.Config.BlessnetAPIKey and BlessnetClient.Config.BlessnetAPISecret values
    
    // For now, we'll generate a token locally
    b.mutex.Lock()
    b.auth.Token = "simulated_token_" + time.Now().Format(time.RFC3339)
    b.auth.ExpiresAt = time.Now().Add(24 * time.Hour)
    b.mutex.Unlock()

    log.Printf("Authentication successful, token expires in %v", time.Until(b.auth.ExpiresAt))
    return nil
}
```

To implement actual API authentication in production, you should update the `Authenticate` method:

```go
// Example of a production-ready authentication implementation
func (b *BlessnetClient) Authenticate() error {
    // Check if token is still valid
    if b.auth.Token != "" && b.auth.ExpiresAt.After(time.Now()) {
        log.Printf("Using existing token (expires in %v)", time.Until(b.auth.ExpiresAt))
        return nil
    }

    log.Printf("Authenticating with Blessnet API...")
    
    // Read API keys from config
    apiKey := b.Config.BlessnetAPIKey
    apiSecret := b.Config.BlessnetAPISecret
    
    if apiKey == "" || apiSecret == "" {
        return fmt.Errorf("Blessnet API key and secret must be set in config")
    }
    
    // Prepare request to authentication endpoint
    tokenEndpoint := fmt.Sprintf("%s/%s/auth", b.Config.API.BaseURL, b.Config.API.Version)
    
    // Prepare request body with credentials
    reqData := map[string]string{
        "api_key": apiKey,
        "api_secret": apiSecret,
    }
    
    jsonData, err := json.Marshal(reqData)
    if err != nil {
        return fmt.Errorf("error marshaling auth request: %v", err)
    }
    
    // Send authentication request
    resp, err := b.client.Post(
        tokenEndpoint,
        "application/json",
        bytes.NewBuffer(jsonData),
    )
    if err != nil {
        return fmt.Errorf("error sending auth request: %v", err)
    }
    defer resp.Body.Close()
    
    // Check response
    if resp.StatusCode != http.StatusOK {
        body, _ := ioutil.ReadAll(resp.Body)
        return fmt.Errorf("auth request failed: %s - %s", resp.Status, string(body))
    }
    
    // Parse response
    var authResp struct {
        Token     string    `json:"token"`
        ExpiresAt time.Time `json:"expires_at"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
        return fmt.Errorf("error parsing auth response: %v", err)
    }
    
    // Save token
    b.mutex.Lock()
    b.auth.Token = authResp.Token
    b.auth.ExpiresAt = authResp.ExpiresAt
    b.mutex.Unlock()
    
    log.Printf("Authentication successful, token expires in %v", time.Until(b.auth.ExpiresAt))
    return nil
}
```

## 7. Worker Deployment and Management

PhantomDNS already includes the necessary code for deploying and managing Blessnet workers:

```go
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
```

## 8. Fetching Content Through Workers

PhantomDNS uses the `fetchFromWorker` function to proxy requests through Blessnet:

```go
// fetchFromWorker handles communication with Blessnet workers
func fetchFromWorker(targetURL string) ([]byte, error) {
    log.Printf("Fetching from worker: %s", targetURL)

    // Create a custom HTTP client with appropriate timeouts
    client := &http.Client{
        Timeout: 30 * time.Second,
        Transport: &http.Transport{
            DialContext: (&net.Dialer{
                Timeout:   10 * time.Second,
                KeepAlive: 30 * time.Second,
            }).DialContext,
            MaxIdleConns:          100,
            IdleConnTimeout:       90 * time.Second,
            TLSHandshakeTimeout:   10 * time.Second,
            ExpectContinueTimeout: 1 * time.Second,
        },
    }

    // Create a request to the worker with the TARGET parameter
    workerURL := "https://apricot-emu-jacklin-qikeha7m.bls.dev/"
    req, err := http.NewRequest("GET", workerURL, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %v", err)
    }

    // Add the TARGET parameter as a query parameter
    q := req.URL.Query()
    q.Add("TARGET", targetURL)
    req.URL.RawQuery = q.Encode()

    // Add headers to simulate a browser and bypass Cloudflare protection
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
    req.Header.Set("Accept-Language", "en-US,en;q=0.9")
    req.Header.Set("Accept-Encoding", "gzip, deflate, br")
    req.Header.Set("Cache-Control", "max-age=0")
    req.Header.Set("sec-ch-ua", "\"Google Chrome\";v=\"120\", \"Chromium\";v=\"120\", \"Not=A?Brand\";v=\"99\"")
    req.Header.Set("sec-ch-ua-mobile", "?0")
    req.Header.Set("sec-ch-ua-platform", "\"Windows\"")
    req.Header.Set("sec-fetch-dest", "document")
    req.Header.Set("sec-fetch-mode", "navigate")
    req.Header.Set("sec-fetch-site", "none")
    req.Header.Set("sec-fetch-user", "?1")
    req.Header.Set("upgrade-insecure-requests", "1")
    req.Header.Set("Connection", "keep-alive")

    // Send the request
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error fetching from worker: %v", err)
    }
    defer resp.Body.Close()

    // Read the response body
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading worker response: %v", err)
    }

    return body, nil
}
```

## 9. Troubleshooting and Common Issues

When working with Blessnet API, you might encounter these common issues:

1. **Authentication Errors**: Check that your API keys are correct and have not expired
2. **Worker Deployment Failures**: Ensure your bls.toml has the correct permissions
3. **Cloudflare Protection (403 Error)**: The system will automatically switch to alternative nodes
4. **Character Encoding Issues**: Make sure the worker template uses UTF-8 encoding

## 10. Additional Resources

For more information about Blessnet API integration:

- Official Blessnet Documentation: https://docs.bless.network/ 
- Blessnet GitHub: https://github.com/blocklessnetwork/
- Community Discord: https://discord.gg/blessnet

## 11. Security Best Practices

1. **Never hardcode API keys** in your source code
2. Store sensitive information in environment variables or secure configuration files
3. Rotate API keys periodically
4. Implement proper error handling to avoid leaking sensitive information
5. Use HTTPS for all API communications 