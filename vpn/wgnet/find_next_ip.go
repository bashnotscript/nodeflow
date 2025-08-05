// internal/wgnet/find_next_ip.go
package wgnet

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strings"
    "myvpn/internal/config"
)

func findNextIP(cfg config.Config) (net.IPNet, error) {
    confPath := fmt.Sprintf("/etc/wireguard/%s.conf", cfg.Interface)
    f, err := os.Open(confPath)
    if err != nil {
        return net.IPNet{}, fmt.Errorf("failed to open config: %w", err)
    }
    defer f.Close()

    usedIPs := make(map[string]bool)
    scanner := bufio.NewScanner(f)
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if strings.HasPrefix(line, "AllowedIPs") {
            parts := strings.SplitN(line, "=", 2)
            if len(parts) == 2 {
                ip := strings.TrimSpace(parts[1])
                usedIPs[ip] = true
            }
        }
    }

    // Scan the subnet to find next available
    ip := cfg.Subnet.IP.Mask(cfg.Subnet.Mask)
    for i := 2; i < 255; i++ {
        candidate := net.IPv4(ip[0], ip[1], ip[2], byte(i))
        cidr := fmt.Sprintf("%s/%d", candidate.String(), maskSize(cfg.Subnet.Mask))
        if !usedIPs[cidr] {
            return net.IPNet{
                IP:   candidate,
                Mask: cfg.Subnet.Mask,
            }, nil
        }
    }

    return net.IPNet{}, fmt.Errorf("no available IPs in subnet")
}

func maskSize(mask net.IPMask) int {
    ones, _ := mask.Size()
    return ones
}

