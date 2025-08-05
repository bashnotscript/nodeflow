// cmd/server.go
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"syscall"

	"nodeflow/vpn/config"
	"nodeflow/vpn/server"
	"nodeflow/vpn/wgnet"
)

func main() {
	iface := flag.String("iface", "wg0", "WireGuard interface name")
	apiPort := flag.Int("api-port", 51820, "API server port")
	token := flag.String("token", "", "API authentication token")
	daemon := flag.Bool("daemon", false, "Run server in background (daemon mode)")
	flag.Parse()

	if *token == "" {
		log.Fatal("Error: --token is required")
	}

	// Daemon mode: fork and run in background
	if *daemon {
		cmd := exec.Command(os.Args[0], append(os.Args[1:], "--daemon=false")...)
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

		if err := cmd.Start(); err != nil {
			log.Fatalf("Failed to start daemon: %v", err)
		}
		fmt.Printf("Server running in background with PID %d\n", cmd.Process.Pid)
		os.Exit(0)
	}

	// Load config from <iface>.conf
	cfg, err := config.LoadFromFile(*iface)
	if err != nil {
		log.Printf("No existing config for %s, creating interface...", *iface)
		_, ipNet, err := net.ParseCIDR("10.0.0.1/24")
		if err != nil {
			log.Fatalf("Failed to parse default CIDR: %v", err)
		}
		pubKey, err := wgnet.DiscoverOrCreateInterface(*iface, ipNet, *apiPort)
		if err != nil {
			log.Fatalf("Failed to create interface: %v", err)
		}
		cfg = &config.Config{
			IfaceName: *iface,
			ListenPort: *apiPort,
			VPNCIDR: ipNet,
			Token: *token,
			PublicKey: pubKey,
		}
	} else {
		log.Printf("Loaded existing config for %s", *iface)
		cfg.IfaceName = *iface
		cfg.ListenPort = *apiPort
		cfg.Token = *token
	}

	// Start HTTP API
	mux := http.NewServeMux()
	mux.Handle("/join", server.JoinHandler(*cfg))

	addr := fmt.Sprintf(":%d", *apiPort)
	log.Printf("API server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("API server error: %v", err)
	}
}

