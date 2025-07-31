package main

import (
    "flag"
    "fmt"
    "log"
    "net"
    "net/http"
    "os"
	  "vpn/wgent"

    "golang.zx2c4.com/wireguard/wgctrl"
)

type Config struct {
    IfaceName  string
    ListenPort int
    VPNCIDR    *net.IPNet
    Token      string
}

func main() {
    // --- Command-Line Flags ---
    iface := flag.String("iface", "wg0", "WireGuard interface name")
    address := flag.String("address", "10.8.0.1/24", "IP address range to assign (CIDR)")
    port := flag.Int("port", 51820, "Listen port for WireGuard")
    token := flag.String("token", "", "API token for authenticating join requests")

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        flag.PrintDefaults()
    }
    flag.Parse()

    if *token == "" {
        log.Fatal("❌ Error: -token flag is required.")
    }

    // --- Parse Network ---
    _, vpnCIDR, err := net.ParseCIDR(*address)
    if err != nil {
        log.Fatalf("Failed to parse CIDR from address %s: %v", *address, err)
    }

    cfg := Config{
        IfaceName:  *iface,
        ListenPort: *port,
        VPNCIDR:    vpnCIDR,
        Token:      *token,
    }

    // --- Discover or Create Interface ---
    log.Println("🔧 Starting WireGuard Mesh Server...")
    wgClient, err := wgctrl.New()
    if err != nil {
        log.Fatalf("Failed to open wgctrl: %v", err)
    }
    defer wgClient.Close()

    if err := wgnet.DiscoverOrCreateInterface(cfg); err != nil {
        log.Fatalf("Interface setup failed: %v", err)
    }

    // --- Start API Server ---
    http.HandleFunc("/join", JoinHandler(cfg)) // pass config to handler
    log.Printf("✅ API server listening on :8080. Use token: %s", cfg.Token)
    log.Fatal(http.ListenAndServe(":8080", nil))
}

