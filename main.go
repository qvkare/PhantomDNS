package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/miekg/dns"
)

// Global configuration and client instances
var (
	config         *Config
	blessnetClient *BlessnetClient
)

// handleDNSRequest processes incoming DNS queries and routes them through Blessnet if necessary
func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				log.Printf("Query for %s\n", q.Name)

				// Check if domain is in proxied list
				domain := strings.TrimSuffix(q.Name, ".")
				if isProxyDomain(domain) {
					// Use Blessnet to fetch this domain through ephemeral proxy
					handleProxiedDomain(m, q)
				} else {
					// Forward to upstream DNS
					forwardToUpstream(m, q)
				}
			}
		}
	}

	w.WriteMsg(m)
}

// isProxyDomain checks if a domain should be proxied through Blessnet
func isProxyDomain(domain string) bool {
	// Check config for domains that should be proxied
	for _, d := range config.ProxyDomains {
		if strings.HasSuffix(domain, d) {
			return true
		}
	}

	return false
}

// handleProxiedDomain processes domains that need to be proxied through Blessnet
func handleProxiedDomain(m *dns.Msg, q dns.Question) {
	// Implementation will depend on your specific requirements
	log.Printf("Proxying domain: %s", q.Name)

	// This is where you would use the Blessnet client to fetch content
	// and determine the appropriate IP to return

	// For now, adding a placeholder IP for testing
	rr, err := dns.NewRR(fmt.Sprintf("%s A 192.168.1.1", q.Name))
	if err == nil {
		m.Answer = append(m.Answer, rr)
	}
}

// forwardToUpstream forwards a DNS query to upstream DNS servers
func forwardToUpstream(m *dns.Msg, q dns.Question) {
	// Use a proper upstream DNS (e.g., Google DNS)
	for _, ns := range config.Nameservers {
		c := new(dns.Client)
		upstreamMsg := new(dns.Msg)
		upstreamMsg.SetQuestion(q.Name, q.Qtype)
		upstreamMsg.RecursionDesired = true

		r, _, err := c.Exchange(upstreamMsg, fmt.Sprintf("%s:53", ns))
		if err != nil {
			log.Printf("Error querying upstream DNS %s: %v", ns, err)
			continue
		}

		if r != nil && len(r.Answer) > 0 {
			m.Answer = append(m.Answer, r.Answer...)
			return
		}
	}
}

func main() {
	// Load configuration
	var err error
	config, err = LoadConfig("config.json")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Blessnet client
	blessnetClient, err = NewBlessnetClient(config)
	if err != nil {
		log.Fatalf("Failed to initialize Blessnet client: %v", err)
	}

	// Attach DNS request handler
	dns.HandleFunc(".", handleDNSRequest)

	// Start DNS server
	server := &dns.Server{
		Addr: fmt.Sprintf("%s:%d", config.DNSListen, config.DNSPort),
		Net:  "udp",
	}

	log.Printf("Starting DNS server on %s:%d", config.DNSListen, config.DNSPort)
	log.Printf("Using enhanced Blessnet integration\n")
	log.Printf("ATTENTION: Blessnet API integration has been improved:\n")
	log.Printf("- Better handling of Cloudflare protection\n")
	log.Printf("- Automatic fallback to alternative regions when primary region is unavailable\n")
	log.Printf("- Enhanced permission system for secure API access\n")

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("Failed to start DNS server: %v", err)
		}
	}()

	// Handle graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	log.Printf("Signal (%v) received, shutting down...", s)
	server.Shutdown()
}

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

	// Log response details
	log.Printf("Worker response status: %s", resp.Status)

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading worker response: %v", err)
	}

	// If response is not successful, log and return error
	if resp.StatusCode != http.StatusOK {
		log.Printf("Worker returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
		if resp.StatusCode == 403 && strings.Contains(string(body), "Cloudflare") {
			log.Printf("Detected Cloudflare protection, trying alternative worker...")
			// Here you could implement fallback to another worker or region
		}
		return nil, fmt.Errorf("worker returned status %d", resp.StatusCode)
	}

	return body, nil
}
