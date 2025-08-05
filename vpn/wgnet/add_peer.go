package wgnet

import (
    "fmt"
    "net"
    "os"

    "nodeflow/vpn/config"
    "golang.zx2c4.com/wireguard/wgctrl"
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func AddPeer(cfg config.Config, peer config.PeerConfig) error {
    client, err := wgctrl.New()
    if err != nil {
        return fmt.Errorf("failed to create wg client: %w", err)
    }
    defer client.Close()

    // Parse AllowedIP (supports string from LoadFromFile or already-net.IPNet)
    var allowedNet net.IPNet
    switch v := any(peer.AllowedIP).(type) {
    case string:
        ip, ipNet, err := net.ParseCIDR(v)
        if err != nil {
            return fmt.Errorf("invalid AllowedIP CIDR %q: %w", v, err)
        }
        allowedNet = *ipNet
        allowedNet.IP = ip
    case net.IPNet:
        allowedNet = v
    default:
        return fmt.Errorf("unsupported AllowedIP type: %T", peer.AllowedIP)
    }

    // Apply to running interface
    wgPeer := wgtypes.PeerConfig{
        PublicKey:         peer.PublicKey,
        AllowedIPs:        []net.IPNet{allowedNet},
        ReplaceAllowedIPs: true,
    }

    if err := client.ConfigureDevice(cfg.IfaceName, wgtypes.Config{
        Peers: []wgtypes.PeerConfig{wgPeer},
    }); err != nil {
        return fmt.Errorf("failed to configure device: %w", err)
    }

    // Persist to <iface>.conf in LoadFromFile-compatible format
    confBlock := fmt.Sprintf(
        "\n[Peer]\nPublicKey = %s\nAllowedIPs = %s\n",
        peer.PublicKey.String(),
        allowedNet.String(),
    )

    confPath := fmt.Sprintf("/etc/wireguard/%s.conf", cfg.IfaceName)
    f, err := os.OpenFile(confPath, os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        return fmt.Errorf("failed to open config file for append: %w", err)
    }
    defer f.Close()

    if _, err := f.WriteString(confBlock); err != nil {
        return fmt.Errorf("failed to append peer to config file: %w", err)
    }

    return nil
}

