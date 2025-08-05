# NodeFlow 
Nodeflow is a self-hosted peer-to-peer VPN coordination tool built on WireGuard, using native Linux Netlink and wgctrl for kernel-level interface configuration. It allows fully decentralized tunnel setups where every peer can act as a server or client within multiple tunnels. Designed for secure, token-based dynamic peer management — no need for centralized databases or JSON config files.

## Features

- Fast and secure VPN via WireGuard
- Native kernel configuration using Netlink and wgctrl (no external tools like wg-quick)
- Token-based authentication for peer joins
- Each peer can act as a server or client in different tunnels
- No external persistence — interface state read directly from kernel
- Modular design: server and agent with shared networking logic

## Server Prerequisites
Linux OS (Ubuntu, Debian, CentOS, or similar) — The server is expected to run on a Linux system with root privileges.
WireGuard kernel module installed and loaded
Go runtime (if building from source) — version 1.18 or newer
Root privileges to create network interfaces and assign IPs.
Chmod 0600 permissions to the /etc/wireguard folder
Ensure port 51820/UDP (or your configured port) is open in your firewall.
The server will create and manage a WireGuard interface (default wg0).

## Quick Start

- Build the tool

```go
git clone https://github.com/bashnotscript/nodeflow.git
cd nodeflow
go build ./cli/server
go build ./cli/agent
```

- Start the Server

```go

./server \
--iface wg0 \
--port 51820 \
--address 10.10.0.0/24 \
--token my-secure-token
# This sets up the WireGuard interface, assigns the first IP, and starts the listener for peer joins.
```

- Join from a Peer (Agent)

```go
./agent \
--server [http://your-server-ip:8080](http://your-server-ip:8080/) \
--token my-secure-token \
--iface wg0
# This fetches configuration from the server and creates a local WireGuard interface accordingly.
```

## Security

Joins are only accepted via valid pre-shared tokens.
AllowedIPs on each peer is strictly scoped to their assigned IP.
Configuration is stored via wgctrl and netlink, avoiding insecure local storage.

## Dependencies

Go 1.20+
Linux (uses Netlink under the hood)
WireGuard tools (kernel module must be present)

## Roadmap

- [ ]  NAT traversal for seamless peer-to-peer communication
- [ ]  Peer list sync across trusted servers
- [ ]  Optional REST admin UI
- [ ]  Expiring tokens and join rate-limiting
- [ ]  Secure VNC/RDP access between nodes

## Contributing

Pull requests are welcome. Please open an issue first to discuss what you would like to change.
