package wgnet

import (
    "fmt"
    "math/rand"
    "net"
    "time"

    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func FindNextIP(vpnCIDR *net.IPNet, currentPeers []wgtypes.Peer) (net.IP, error) {
    usedIPs := make(map[string]bool)
    usedIPs[vpnCIDR.IP.String()] = true // exclude network address

    for _, peer := range currentPeers {
        for _, allowed := range peer.AllowedIPs {
            usedIPs[allowed.IP.String()] = true
        }
    }

    // convert subnet mask to prefix length
    prefixLen, _ := vpnCIDR.Mask.Size()
    networkBase := vpnCIDR.IP.Mask(vpnCIDR.Mask).To4()

    // Iterate IPs within subnet, skipping network and broadcast addresses
    start := ipToInt(networkBase) + 1
    end := start | ((1 << (32 - prefixLen)) - 2) // last usable IP

	  r := rand.New(rand.NewSource(time.Now().UnixNano()))
    tries := 100

    for range tries {
        candidateInt := start + uint32(rand.Intn(r.Intn(int(end-start)+1)))
        candidateIP := intToIP(candidateInt)

        if !usedIPs[candidateIP.String()] && vpnCIDR.Contains(candidateIP) {
            return candidateIP, nil
        }
    }

    return nil, fmt.Errorf("no available IP found after %d tries", tries)
}

func ipToInt(ip net.IP) uint32 {
    ip = ip.To4()
    return uint32(ip[0])<<24 + uint32(ip[1])<<16 + uint32(ip[2])<<8 + uint32(ip[3])
}

func intToIP(i uint32) net.IP {
    return net.IPv4(byte(i>>24), byte(i>>16), byte(i>>8), byte(i))
}

