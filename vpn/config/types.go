package config

import (
    "net"

    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type Config struct {
    IfaceName  string
    ListenPort int
    VPNCIDR    *net.IPNet
    Token      string
    PublicKey  wgtypes.Key
}

type PeerConfig struct {
    PublicKey wgtypes.Key
    AllowedIP string // CIDR string, e.g. "10.0.0.2/32"
}

type JoinRequest struct {
    APIToken  string `json:"api_token"`
    PublicKey string `json:"public_key"`
}

type JoinResponse struct {
    AssignedIP      string          `json:"assigned_ip"`
    ServerPublicKey string          `json:"server_public_key"`
    Peers           []wgtypes.Peer  `json:"peers"`
}

