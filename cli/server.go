package main

import (
    "flag"
    "log"
    "net"
    "net/http"
    "os"
    "strings"

    "nodeflow/vpn/config"
    "nodeflow/vpn/server"
    "nodeflow/vpn/wgnet"
)

func main() {
    iface := flag.String("iface", "wg0", "WireGuard interface name")
    cidrStr := flag.String("cidr", "10.0.0.1/24", "VPN CIDR")
    listenPort := flag.Int("port", 51820, "WireGuard listen port")
    apiAddr := flag.String("api", ":8080", "API listen address")
    token := flag.String("token", "", "API join token (required)")
    flag.Parse()

    if *token == "" {
        log.Fatal("API token is required via -token")
    }

    // Parse VPN CIDR
    _, vpnCIDR, err := net.ParseCIDR(*cidrStr)
    if err != nil {
        log.Fatalf("Invalid CIDR %s: %v", *cidrStr, err)
    }

    confPath := "/etc/wireguard/" + *iface + ".conf"
    var cfg config.Config

    // If config exists, load it — otherwise create it
    if _, err := os.Stat(confPath); err == nil {
        cfg, err = config.LoadFromFile(confPath)
        if err != nil {
            log.Fatalf("Failed to load config: %v", err)
        }
        log.Printf("Loaded existing config from %s", confPath)
    } else {
        // Discover or create interface
        pubKey, err := wgnet.DiscoverOrCreateInterface(*iface, vpnCIDR, *listenPort)
        if err != nil {
            log.Fatalf("Failed to create or find interface: %v", err)
        }

        // Store initial config
        cfg = config.Config{
            InterfaceName: *iface,
            VPNCIDR:       vpnCIDR,
            ListenPort:    *listenPort,
            Token:         *token,
            PrivateKey:    "", // Already in .conf, no need to store in struct for security
            Peers:         []config.PeerConfig{},
        }

        log.Printf("Interface %s ready. Public key: %s", *iface, pubKey.String())
    }

    // Ensure token matches the one from CLI if loaded from file
    if strings.TrimSpace(cfg.Token) == "" {
        cfg.Token = *token
    }

    // HTTP handler for join
    http.Handle("/join", server.JoinHandler(cfg))

    log.Printf("API listening on %s", *apiAddr)
    log.Fatal(http.ListenAndServe(*apiAddr, nil))
}

