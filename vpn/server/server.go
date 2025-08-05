package server

import (
    "encoding/json"
    "net/http"
	   "net"

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
		
		            // Get current peers
        wgClient, err := wgctrl.New()
        if err != nil {
            http.Error(w, "Failed to open wgctrl", http.StatusInternalServerError)
            return
        }
        defer wgClient.Close()

        device, err := wgClient.Device(cfg.IfaceName)
        if err != nil {
            http.Error(w, "Failed to get WG device", http.StatusInternalServerError)
            return
        }
        nextIP, err := wgnet.FindNextIP(cfg.VPNCIDR, device.Peers)	
        if err != nil {
            http.Error(w, "IP allocation failed", http.StatusInternalServerError)
            return
        }
        peerIPNet := net.IPNet{
            IP:   nextIP,
            Mask: net.CIDRMask(32, 32),
        }


        peer := config.PeerConfig{
            PublicKey: pubKey,
            AllowedIP: peerIPNet.String(),
        }

        if err := wgnet.AddPeer(cfg, peer); err != nil {
            http.Error(w, "Failed to add peer", http.StatusInternalServerError)
            return
        }
		    
		            // Refresh device to get updated peers
        device, err = wgClient.Device(cfg.IfaceName)
        if err != nil {
            http.Error(w, "Failed to refresh WG device", http.StatusInternalServerError)
            return
        }

        // Prepare peer list for response
        //peers := config.peerDetails{}
        //for _, p := range device.Peers {
          //  for _, allowedIP := range p.AllowedIPs {
            //    peers = append(peers, peerDetail{
              //      PublicKey:  p.PublicKey.String(),
                //    AllowedIPs: allowedIP.String(),
               // })
           // }
       // }

        resp := config.JoinResponse{
            AssignedIP:      nextIP.String(),
            ServerPublicKey: device.PublicKey.String(),
            Peers:           device.Peers,
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(resp)
	

    }
}


