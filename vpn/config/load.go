package config

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"

    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func LoadFromFile(iface string) (*Config, error) {
    path := fmt.Sprintf("/etc/wireguard/%s.conf", iface)
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var cfg Config
    cfg.IfaceName = iface

    scanner := bufio.NewScanner(file)
    var inInterface bool

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        switch line {
        case "[Interface]":
            inInterface = true
        case "[Peer]":
            inInterface = false
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
                    pk, err := wgtypes.ParseKey(val)
                    if err == nil {
                        cfg.PublicKey = pk.PublicKey()
                    }
                case "Address":
                    _, ipnet, err := net.ParseCIDR(val)
                    if err == nil {
                        cfg.VPNCIDR = ipnet
                    }
                case "ListenPort":
                    fmt.Sscanf(val, "%d", &cfg.ListenPort)
                }
            }
        }
    }
    return &cfg, nil
}

