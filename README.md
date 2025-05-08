# PhantomDNS

PhantomDNS is a DNS-based censorship circumvention tool. It uses the Blessnet distributed network to bypass fixed IP/SNI signatures and DNS blocks by utilizing ephemeral proxy workers that operate from different IPs and locations for each HTTP request.

**Current Worker URL:** [https://apricot-emu-jacklin-qikeha7m.bls.dev/](https://apricot-emu-jacklin-qikeha7m.bls.dev/)

## Features

- **DNS-based Censorship Circumvention**: PhantomDNS runs as a local DNS server to intercept requests to blocked domains.
- **Ephemeral Proxy Workers**: Each request gets a fresh worker with a different IP and geolocation.
- **Distributed Network**: Leverages the Blessnet network to distribute requests across multiple regions.
- **No Logging**: Workers are designed for privacy with minimal logging and short lifetimes.
- **Easy Configuration**: Simple JSON configuration for proxy domains and DNS settings.

## Installation

### Requirements

- Go 1.18 or higher
- Node.js and npm (for Blessnet CLI)
- Blessnet CLI: `npm install -g blessnet-cli`

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/qvkare/PhantomDNS.git
cd PhantomDns

# Install Go dependencies
go mod download

# Install Blessnet worker dependencies
npm install
```

### Configuration

Create a `config.json` file with your settings:

```json
{
  "dns_port": 5355,
  "dns_listen": "127.0.0.1",
  "nameservers": ["8.8.8.8", "1.1.1.1"],
  "blessnet_worker_url": "https://your-worker-url.bls.dev",
  "proxy_domains": ["blocked.com", "restricted.org"],
  "proxy_mode": "ephemeral"
}
```

## Usage

### Starting the DNS Server

```bash
# Using the shell script (recommended)
./run.sh

# Or manually with Go
go run main.go blessnet.go blessnet_api.go config.go
```

### Client Configuration

Configure your system or applications to use PhantomDNS as the DNS server:

#### Testing with dig

```bash
dig @127.0.0.1 -p 5355 blocked.com
```

#### Configure System DNS

**Linux/macOS**:
Edit `/etc/resolv.conf` or use Network Manager to set DNS to `127.0.0.1` with port `5355`.

**Windows**:
Set DNS server to `127.0.0.1` in network adapter settings.

### Browser Configuration

Most browsers respect system DNS settings, but you can also:

- Use a browser extension that allows custom DNS settings
- Use a proxy extension configured to use PhantomDNS

## Blessnet Integration

PhantomDNS leverages the Blessnet distributed network for its ephemeral workers. To configure Blessnet:

```bash
# Configure Blessnet CLI
blessnet manage

# Deploy workers
blessnet deploy
```

## Security Notes

- PhantomDNS is designed for privacy, but is not a complete security solution
- Always use HTTPS for sensitive communications
- Consider additional security layers for high-risk situations

## Development

### Project Structure

- `main.go` - DNS server and main application logic
- `blessnet.go` - Blessnet client implementation
- `config.go` - Configuration handling
- `blessnet_api.go` - Blessnet API interactions
- `src/index.ts` - Worker code for Blessnet

### Building from Source

```bash
# Build for current platform
go build -o phantomdns main.go blessnet.go blessnet_api.go config.go

# Cross-compile for other platforms
GOOS=windows GOARCH=amd64 go build -o phantomdns.exe main.go blessnet.go blessnet_api.go config.go
```

## License

[MIT License](LICENSE)

## Contributing

Contributions welcome! Please feel free to submit a Pull Request.

## Acknowledgements

- This project utilizes the [Blessnet](https://bless.network/) distributed network
- Thanks to all contributors and the censorship circumvention community