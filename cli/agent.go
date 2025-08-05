package main

import (
    "flag"
    "log"
    "nodeflow/vpn/agent"
)

func main() {
    serverURL := flag.String("server", "", "The full URL of the server's API endpoint (e.g., http://192.168.33.10:8080/join)")
    apiToken := flag.String("token", "", "The secret API token to join the network")
    iface := flag.String("iface", "wg0", "Name for the local WireGuard interface")
    allowedIPs := flag.String("allowed", "0.0.0.0/0", "Comma-separated list of allowed IPs for the WireGuard peer")
    flag.Parse()

    if *serverURL == "" || *apiToken == "" {
        log.Fatal("Usage: agent -server=<url> -token=<token> [-iface=<iface>] [-allowed=<ips>]")
    }

    err := agent.JoinNetwork(*serverURL, *apiToken, *iface, *allowedIPs)
    if err != nil {
        log.Fatalf("Failed to join network: %v", err)
    }
}

