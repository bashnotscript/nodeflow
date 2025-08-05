// internal/wgnet/add_peer.go
package wgnet

import (
    "fmt"
    "net"
    "os"
    "strings"
    "myvpn/internal/config"

    "golang.zx2c4.com/wireguard/wgctrl"
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func AddPeer(cfg config.Config, peer config.PeerConfig) error {
    client, err := wgctrl.New()
    if err != nil {
        return fmt.Errorf("failed to create wg client: %w", err)
    }
    defer client.Close()

    allowedIP := net.IPNet{
        IP:   peer.AllowedIP.IP,
        Mask: peer.AllowedIP.Mask,
    }

    wgPeer := wgtypes.PeerConfig{
        PublicKey:         peer.PublicKey,
        AllowedIPs:        []net.IPNet{allowedIP},
        ReplaceAllowedIPs: true,
    }

    if err := client.ConfigureDevice(cfg.Interface, wgtypes.Config{
        Peers: []wgtypes.PeerConfig{wgPeer},
    }); err != nil {
        return fmt.Errorf("failed to configure device: %w", err)
    }

    // Persist to <iface>.conf
    confLine := fmt.Sprintf("\n[Peer]\nPublicKey = %s\nAllowedIPs = %s\n",
        peer.PublicKey.String(), allowedIP.String(),
    )

    confPath := fmt.Sprintf("/etc/wireguard/%s.conf", cfg.Interface)
    f, err := os.OpenFile(confPath, os.O_APPEND|os.O_WRONLY, 0600)
    if err != nil {
        return fmt.Errorf("failed to open config file: %w", err)
    }
    defer f.Close()

    if _, err := f.WriteString(confLine); err != nil {
        return fmt.Errorf("failed to write to config file: %w", err)
    }

    return nil
}

