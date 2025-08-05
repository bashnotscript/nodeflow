package server

import (
    "encoding/json"
    "net/http"

    "golang.zx2c4.com/wireguard/wgctrl"
    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

	  "nodeflow/vpn/config"	
		"nodeflow/vpn/wgnet"
)


func JoinHandler(cfg config.Config) http.HandlerFunc {
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

        nextIP, err := FindNextIP(cfg.VPNCIDR, device.Peers)
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


