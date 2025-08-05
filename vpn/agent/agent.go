package agent

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strings"

    "golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

type joinRequest struct {
    PublicKey string `json:"public_key"`
}

type joinResponse struct {
    AssignedIP      string       `json:"assigned_ip"`
    ServerPublicKey string       `json:"server_public_key"`
    Peers           []peerDetail `json:"peers"`
}

type peerDetail struct {
    PublicKey  string `json:"public_key"`
    AllowedIPs string `json:"allowed_ip"`
}

// JoinNetwork joins the WireGuard mesh network by requesting an IP and peers from the server.
func JoinNetwork(serverURL, apiToken, iface, allowedIPs string) error {
    // Generate client keypair
    privKey, err := wgtypes.GeneratePrivateKey()
    if err != nil {
        return fmt.Errorf("failed to generate private key: %w", err)
    }
    pubKey := privKey.PublicKey()

    // Prepare join request
    reqPayload := joinRequest{
        PublicKey: pubKey.String(),
    }
    reqBytes, err := json.Marshal(reqPayload)
    if err != nil {
        return fmt.Errorf("failed to marshal join request: %w", err)
    }

    // Create request with token header
    req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(reqBytes))
    if err != nil {
        return fmt.Errorf("failed to create request: %w", err)
    }
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Token", apiToken)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
    }

    var jr joinResponse
    if err := json.NewDecoder(resp.Body).Decode(&jr); err != nil {
        return fmt.Errorf("failed to decode response: %w", err)
    }

    log.Printf("✅ Joined network! Assigned IP: %s", jr.AssignedIP)

    // Generate config file content
    configContent := buildConfig(privKey.String(), jr, iface, allowedIPs)

    fileName := fmt.Sprintf("%s.conf", iface)
    if err := os.WriteFile(fileName, []byte(configContent), 0600); err != nil {
        return fmt.Errorf("failed to write config file: %w", err)
    }
    log.Printf("✅ WireGuard config saved to %s", fileName)

    // Print config summary for user convenience
    printConfigSummary(privKey.String(), jr, allowedIPs)

    return nil
}

func buildConfig(privKey string, jr joinResponse, iface, allowedIPs string) string {
    b := &strings.Builder{}
    b.WriteString("[Interface]\n")
    b.WriteString(fmt.Sprintf("PrivateKey = %s\n", privKey))
    b.WriteString(fmt.Sprintf("Address = %s\n\n", jr.AssignedIP))

    b.WriteString("[Peer]\n")
    b.WriteString(fmt.Sprintf("PublicKey = %s\n", jr.ServerPublicKey))
    b.WriteString(fmt.Sprintf("AllowedIPs = %s\n", allowedIPs))
    b.WriteString("PersistentKeepalive = 25\n\n")

    // Add other peers for discovery (optional)
    for _, p := range jr.Peers {
        b.WriteString("[Peer]\n")
        b.WriteString(fmt.Sprintf("PublicKey = %s\n", p.PublicKey))
        b.WriteString(fmt.Sprintf("AllowedIPs = %s\n\n", p.AllowedIPs))
    }

    return b.String()
}

func printConfigSummary(privKey string, jr joinResponse, allowedIPs string) {
    fmt.Println("\n--- WireGuard Client Configuration ---")
    fmt.Println("[Interface]")
    fmt.Printf("PrivateKey = %s\n", privKey)
    fmt.Printf("Address = %s\n\n", jr.AssignedIP)

    fmt.Println("[Peer]")
    fmt.Printf("PublicKey = %s\n", jr.ServerPublicKey)
    fmt.Printf("AllowedIPs = %s\n", allowedIPs)
    fmt.Println("PersistentKeepalive = 25\n")
}

