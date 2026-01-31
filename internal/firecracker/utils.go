package firecracker

import (
	"crypto/md5"
	"fmt"
)

// generateMACAddress generates a deterministic MAC address for a VM based on its ID
// Format: 02:XX:XX:XX:XX:XX (02 = locally administered, unicast)
// Uses MD5 hash of VM ID to ensure deterministic but unique MACs
func generateMACAddress(vmID string) string {
	// Hash VM ID to get deterministic MAC
	hash := md5.Sum([]byte(vmID))
	
	// Use first 5 bytes of hash for MAC address
	// Ensure locally administered bit is set (02:)
	return fmt.Sprintf("02:%02x:%02x:%02x:%02x:%02x", 
		hash[0], hash[1], hash[2], hash[3], hash[4])
}
