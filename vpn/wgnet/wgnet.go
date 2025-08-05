package wgnet

import (
    "fmt"
    "log"
    "net"
    "os"
    "strings"

    "github.com/vishvananda/netlink"
    "golang.zx2c4.com/wireguard/wgctrl"
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// DiscoverOrCreateInterface ensures the WireGuard interface exists, or creates it if not.
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

    // 1. Create WG interface
    link := &netlink.GenericLink{
        LinkAttrs: netlink.LinkAttrs{Name: ifaceName},
        LinkType:  "wireguard",
    }
    if err := netlink.LinkAdd(link); err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to add interface: %w", err)
    }

    // 2. Assign IP address
    addr, err := netlink.ParseAddr(cidr.String())
    if err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to parse CIDR address: %w", err)
    }
    if err := netlink.AddrAdd(link, addr); err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to assign address: %w", err)
    } 
    // 4. Configure device with private key and port
    config := wgtypes.Config{
        PrivateKey:   &privateKey,
        ListenPort:   &listenPort,
        ReplacePeers: true,
        Peers:        []wgtypes.PeerConfig{},
    }
    if err := wgClient.ConfigureDevice(ifaceName, config); err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to configure wireguard device: %w", err)
    }
	
	// 5. Bring interface up
   // if err:= netlink.LinkSetUp(link); err != nil {
		 //   netlink.LinkDel(link)	
       // return wgtypes.Key{}, fmt.Errorf("failed to set interface up: %w", err)
    //}
    // 5. Persist to <iface>.conf (in the format LoadFromFile expects)
    var conf strings.Builder
    conf.WriteString("[Interface]\n")
    conf.WriteString(fmt.Sprintf("PrivateKey = %s\n", privateKey.String()))
    conf.WriteString(fmt.Sprintf("Address = %s\n", cidr.String()))
    conf.WriteString(fmt.Sprintf("ListenPort = %d\n", listenPort))
    conf.WriteString("\n") // Ensure a blank line before any peers get appended

    confPath := fmt.Sprintf("/etc/wireguard/%s.conf", ifaceName)
    if err := os.WriteFile(confPath, []byte(conf.String()), 0600); err != nil {
        return wgtypes.Key{}, fmt.Errorf("failed to write config file: %w", err)
    }

    log.Printf("Created WireGuard config at %s", confPath)
    return privateKey.PublicKey(), nil
}

