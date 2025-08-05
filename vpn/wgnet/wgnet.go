package wgnet

import (
    "fmt"
    "log"
    "net"

    "github.com/vishvananda/netlink"
    "golang.zx2c4.com/wireguard/wgctrl"
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// DiscoverOrCreateInterface ensures the interface exists, or creates it if not.
// Returns the public key used.
func DiscoverOrCreateInterface(ifaceName string, cidr *net.IPNet, listenPort int) (wgtypes.Key, error) {
    wgClient, err := wgctrl.New()
    if err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to open wgctrl: %w", err)
    }
    defer wgClient.Close()

    devices, err := wgClient.Devices()
    if err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to get devices: %w", err)
    }

    for _, dev := range devices {
        if dev.Name == ifaceName {
            log.Printf("Found existing interface '%s'. Using it.", ifaceName)
            return dev.PublicKey, nil
        }
    }

    log.Printf("No interface named '%s' found. Creating a new one.", ifaceName)
    return createInterface(wgClient, ifaceName, cidr, listenPort)
}

func createInterface(wgClient *wgctrl.Client, ifaceName string, cidr *net.IPNet, listenPort int) (wgtypes.Key, error) {
    privateKey, err := wgtypes.GeneratePrivateKey()
    if err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to generate private key: %w", err)
    }

    // 1. Create interface
    link := &netlink.GenericLink{
        LinkAttrs: netlink.LinkAttrs{Name: ifaceName},
        LinkType:  "wireguard",
    }
    if err := netlink.LinkAdd(link); err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to add interface: %w", err)
    }

    // 2. Assign IP
    addr, err := netlink.ParseAddr(cidr.String())
    if err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to parse address: %w", err)
    }
    if err := netlink.AddrAdd(link, addr); err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to add address: %w", err)
    }

    // 3. Bring up
    if err := netlink.LinkSetUp(link); err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to set interface up: %w", err)
    }

    // 4. WireGuard config
    config := wgtypes.Config{
        PrivateKey:   &privateKey,
        ListenPort:   &listenPort,
        ReplacePeers: true,
        Peers:        []wgtypes.PeerConfig{},
    }

    if err := wgClient.ConfigureDevice(ifaceName, config); err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to configure device: %w", err)
    }

    return privateKey.PublicKey(), nil
}

