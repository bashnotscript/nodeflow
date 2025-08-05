package server

import (
    "encoding/binary"
    "encoding/json"
    "fmt"
    "log"
    "math/rand"
    "net"
    "net/http"

    "golang.zx2c4.com/wireguard/wgctrl"
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
    IfaceName  string
    VPNCIDR    *net.IPNet
    Token      string
    PublicKey  wgtypes.Key
}

type JoinRequest struct {
    APIToken  string `json:"api_token"`
    PublicKey string `json:"public_key"`
}

type JoinResponse struct {
    AssignedIP      string           `json:"assigned_ip"`
    ServerPublicKey string           `json:"server_public_key"`
    Peers           []wgtypes.Peer  `json:"peers"`
}

func joinHandler(cfg config.Config) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        if r.Header.Get("X-Token") != cfg.Token {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        var req struct {
            PublicKey string `json:"public_key"`
        }

        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }

        pubKey, err := wgtypes.ParseKey(req.PublicKey)
        if err != nil {
            http.Error(w, "Invalid public key", http.StatusBadRequest)
            return
        }

        nextIP, err := findNextIP(cfg)
        if err != nil {
            http.Error(w, "IP allocation failed", http.StatusInternalServerError)
            return
        }

        peer := config.PeerConfig{
            PublicKey: pubKey,
            AllowedIP: nextIP,
        }

        if err := wgnet.AddPeer(cfg, peer); err != nil {
            http.Error(w, "Failed to add peer", http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]string{
            "ip": nextIP.IP.String(),
        })
    }
}


func findNextIP(cidr *net.IPNet, currentPeers []wgtypes.Peer) (net.IP, error) {
    used := make(map[string]bool)
    used[cidr.IP.String()] = true
    for _, peer := range currentPeers {
        for _, allowed := range peer.AllowedIPs {
            used[allowed.IP.String()] = true
        }
    }

    base := binary.BigEndian.Uint32(cidr.IP.To4())
    mask := binary.BigEndian.Uint32(cidr.Mask)
    start := base + 1
    end := (base & mask) | (mask ^ 0xffffffff)

    for i := 0; i < 100; i++ {
        candidate := start + rand.Uint32()%(end-start-1)
        ip := make(net.IP, 4)
        binary.BigEndian.PutUint32(ip, candidate)
        if !used[ip.String()] {
            return ip, nil
        }
    }

    return nil, fmt.Errorf("could not find an available IP after 100 tries")
}

