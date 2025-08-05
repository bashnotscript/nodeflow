package config

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"

    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// LoadFromFile reads a WireGuard <iface>.conf and fills Config
func LoadFromFile(confPath string) (Config, error) {
    file, err := os.Open(confPath)
    if err != nil {
        return Config{}, err
    }
    defer file.Close()

    var cfg Config
    cfg.Peers = []PeerConfig{}

    scanner := bufio.NewScanner(file)
    var inInterface, inPeer bool
    var peer PeerConfig

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        switch line {
        case "[Interface]":
            inInterface, inPeer = true, false
        case "[Peer]":
            if inPeer && peer.PublicKey != (wgtypes.Key{}) {
                cfg.Peers = append(cfg.Peers, peer)
            }
            peer = PeerConfig{}
            inInterface, inPeer = false, true
        default:
            kv := strings.SplitN(line, "=", 2)
            if len(kv) != 2 {
                continue
            }
            key := strings.TrimSpace(kv[0])
            val := strings.TrimSpace(kv[1])

            if inInterface {
                switch key {
                case "PrivateKey":
                    cfg.PrivateKey = val
                case "Address":
                    ip, ipnet, err := net.ParseCIDR(val)
                    if err == nil {
                        // Ensure IP is the network address
                        ipnet.IP = ip
                        cfg.VPNCIDR = ipnet
                    }
                case "ListenPort":
                    var port int
                    fmt.Sscanf(val, "%d", &port)
                    cfg.ListenPort = port
                }
            } else if inPeer {
                switch key {
                case "PublicKey":
                    pk, err := wgtypes.ParseKey(val)
                    if err == nil {
                        peer.PublicKey = pk
                    }
                case "AllowedIPs":
                    _, ipnet, err := net.ParseCIDR(val)
                    if err == nil {
                        peer.AllowedIP = ipnet
                    }
                }
            }
        }
    }

    if inPeer && peer.PublicKey != (wgtypes.Key{}) {
        cfg.Peers = append(cfg.Peers, peer)
    }

    return cfg, nil
}

