# PhantomDNS User Guide

PhantomDNS is an advanced DNS-based proxy system designed to bypass censorship and site blocking. Unlike traditional VPNs, PhantomDNS uses ephemeral proxy workers that operate with different IP addresses and locations for each request, reducing traceability and bypassing blocks based on fixed IP/SNI signatures.

## Table of Contents

1. [Overview](#overview)
2. [System Requirements](#system-requirements)
3. [Installation](#installation)
4. [Configuration](#configuration)
5. [Usage](#usage)
6. [Troubleshooting](#troubleshooting)
7. [Security and Privacy](#security-and-privacy)
8. [FAQ (Frequently Asked Questions)](#faq)
9. [Technical Details](#technical-details)

## Overview

Here's how PhantomDNS works:

1. PhantomDNS runs as a local DNS server on your computer
2. When you try to access a blocked website, PhantomDNS detects this request
3. The system automatically starts a proxy worker in the Blessnet distributed network
4. This worker accesses the blocked content from a different IP address and location
5. The content is delivered to you, and the worker destroys itself after the transaction

This approach provides a structure that is much harder for censorship systems to detect because the workers used for each request are ephemeral (temporary) and operate from different locations.

## System Requirements

To run PhantomDNS, you need the following:

- Operating System: Windows, macOS, or Linux
- CPU: At least a dual-core processor
- RAM: At least 2GB (4GB recommended)
- Disk: At least 100MB of free space
- Internet: A stable internet connection
- Permissions: Authority to change DNS settings (may require admin rights on some systems)

## Installation

### Windows

1. Download the [PhantomDNS Windows installer](https://phantomdns.org/downloads/windows)
2. Double-click the downloaded file and follow the installation wizard
3. The application will start automatically when the installation is complete
4. You will see the PhantomDNS icon in the system tray

### macOS

1. Download the [PhantomDNS macOS application](https://phantomdns.org/downloads/macos)
2. Open the downloaded DMG file and drag the PhantomDNS application to the Applications folder
3. When you first run the application, you will be asked for permission to change system settings
4. After granting permissions, the application will appear in the menu bar

### Linux

```bash
# Add the repository key
curl -fsSL https://repo.phantomdns.org/key.gpg | sudo apt-key add -

# Add the repository
echo "deb https://repo.phantomdns.org/apt stable main" | sudo tee /etc/apt/sources.list.d/phantomdns.list

# Update package list
sudo apt update

# Install PhantomDNS
sudo apt install phantomdns

# Start the service
sudo systemctl start phantomdns
```

## Configuration

After successfully installing PhantomDNS, you need to configure it:

### General Settings

Open the PhantomDNS application and configure the following settings:

1. **DNS Server Port**: The default is 5354, but you can specify another port
2. **Worker Count**: The number of workers to run simultaneously (default: 3)
3. **Auto Start**: Automatic startup of PhantomDNS when the system boots (recommended)
4. **Log Level**: Set the log level for debugging (for normal use, "Info" is sufficient)

### DNS Settings

You need to configure your system's DNS settings to use PhantomDNS:

#### Windows

1. Open Network Settings
2. Select your active network adapter
3. Properties > IPv4 > Properties
4. Check "Use the following DNS server addresses"
5. Preferred DNS server: `127.0.0.1`
6. Alternate DNS server: `8.8.8.8` (as a backup)
7. Click OK and save the settings

#### macOS

1. System Preferences > Network
2. Select your active connection (Wi-Fi or Ethernet)
3. Advanced > DNS
4. Use the "+" button to add `127.0.0.1` to the DNS Servers list
5. Click "Apply"

#### Linux

Using NetworkManager:

```bash
# Set DNS server to 127.0.0.1
nmcli con mod "Connection Name" ipv4.dns "127.0.0.1"
# Restart the connection
nmcli con up "Connection Name"
```

or edit the `/etc/resolv.conf` file:

```
nameserver 127.0.0.1
nameserver 8.8.8.8  # As a backup
```

## Usage

After installing and configuring PhantomDNS, its usage is completely transparent. Use your web browser normally, and when you try to access blocked websites, PhantomDNS will automatically take over.

### Accessing a Blocked Site

1. Open your web browser
2. Enter the URL of the blocked website (e.g., `https://blocked-site.com`)
3. PhantomDNS works in the background:
   - Detects your DNS request
   - Starts a worker in the Blessnet network
   - The worker accesses the blocked content
   - The content is delivered to you
4. You can view and use the site normally

### Status Check

To check the status of PhantomDNS:

- Windows/macOS: Click on the system tray/menu bar icon
- Linux: Run the `phantomdns status` command from the command line

### Manual Testing

To verify that PhantomDNS is working properly:

```bash
# Windows (Command Prompt)
nslookup blocked-site.com 127.0.0.1 -port=5354

# macOS/Linux (Terminal)
dig @127.0.0.1 -p 5354 blocked-site.com
```

If the request is successful, you will see an IP address generated by PhantomDNS.

## Troubleshooting

Common problems and solutions:

### PhantomDNS Not Working

1. Check if the PhantomDNS service is running
2. Make sure your DNS settings are configured correctly
3. Check firewall settings to ensure the necessary ports for PhantomDNS (5354/UDP, 5354/TCP) are open

### Specific Sites Not Opening

1. Make sure PhantomDNS is up to date
2. Are sites changing rapidly? In some cases, site operators may make continual changes to bypass blocks
3. Try clearing the DNS cache:
   - Windows: `ipconfig /flushdns`
   - macOS: `sudo killall -HUP mDNSResponder`
   - Linux: `sudo systemd-resolve --flush-caches`

### Blessnet Worker Connection Issues

1. **Cloudflare Protection (403 Error)**: Blessnet worker URLs may sometimes be protected by Cloudflare and you might receive a "403 Forbidden" error. In this case:
   - PhantomDNS will switch to the backup mechanism
   - The system automatically switches to alternative nodes
   - Make sure you have enabled alternative regions (such as `eu-west`, `ap-east`) in the `config.json` file

2. **White Page Issue**: If you see a white page when accessing the worker URL (e.g., apricot-emu-jacklin-qikeha7m.bls.dev):
   - There might be a worker initialization error, check the logs: `cat runtime.log`
   - Make sure the worker template is loaded correctly
   - Reconfigure the Blessnet CLI tool: `blessnet manage`
   - Redeploy the worker: `blessnet deploy`

3. **Character Encoding Issues**: If you see character encoding problems in worker responses (like `baÅŸlatÄ±ldÄ±`):
   - Make sure the worker template uses UTF-8 character encoding
   - Add `"Content-Type": "text/plain; charset=utf-8"` to the response header in the `src/index.ts` file
   - After making the changes, recompile and redeploy the worker:
     ```bash
     npm run build:release
     blessnet deploy
     ```

### Slow Connection

1. Try different worker count settings
2. Check your network connection
3. Congestion in the Blessnet network can occasionally affect performance

## Security and Privacy

PhantomDNS is designed to protect your security and privacy:

### Security Features

- **Ephemeral Workers**: New workers are created for each request and destroyed after processing
- **Distributed Network**: Requests are made from different locations through the Blessnet distributed network
- **mTLS Handshake**: The mTLS protocol is used for secure connections
- **No Logging**: Workers do not keep any logs after the transaction

### Privacy Policy

PhantomDNS:
- Does not monitor your internet activity
- Does not collect your personal data
- Does not record your DNS queries
- Does not share data with third parties

## FAQ

### Is PhantomDNS legal?
While the PhantomDNS technology itself is legal, your use of it is subject to the laws of your country. Please check your local legal regulations.

### Is PhantomDNS a VPN?
No, PhantomDNS is not a VPN. While VPNs encrypt all your traffic, PhantomDNS only processes DNS requests and provides access to blocked content.

### Can I use PhantomDNS for free?
A limited free version of PhantomDNS is available, but it is recommended to upgrade to the premium version for full functionality.

### Can it bypass all blocks?
PhantomDNS can bypass most DNS-based blocks and fixed IP/SNI signatures, but its effectiveness may vary in cases where more complex techniques like deep packet inspection (DPI) are used.

## Technical Details

Technical details for advanced users:

### Architecture

PhantomDNS consists of three main components:
1. **DNS Server**: Runs on your local machine and captures DNS requests
2. **Blessnet Workers**: Proxy workers that provide access to blocked sites
3. **Node Management**: Integration layer providing direct API access to Blessnet nodes

### Custom Configuration

You can customize PhantomDNS by editing the `config.yaml` file:

```yaml
listen_address: 127.0.0.1
listen_port: 5354
worker_count: 3
regions:
  - us-east
  - eu-west
  - ap-east
log_level: info
```

### API Integration

If you want to integrate your own applications with PhantomDNS, you can use the local REST API:

```bash
# Status check
curl http://localhost:8017/status

# Start a new worker
curl http://localhost:8017/create_worker?target=https://example.com
```

---

For more information, you can visit the [official website](https://phantomdns.org) or use the [support forum](https://forum.phantomdns.org) for assistance. 