package datapath

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// generateDatapathID generates a datapath ID from the switch name
func GenerateID(switchName string) string {
	// Create a new SHA256 hash object
	hash := sha256.New()

	// Write the switch name to the hash object
	hash.Write([]byte(switchName))

	// Get the hashed bytes
	hashedBytes := hash.Sum(nil)

	// Take the first 8 bytes of the hash to create a 64-bit ID
	dpidBytes := hashedBytes[:8]

	// Convert the bytes to a hexadecimal string
	dpid := hex.EncodeToString(dpidBytes)

	return dpid
}

type DatapathParams struct {
	NodeName     string
	ProviderName string
}

func GetSwitchName(params DatapathParams) string {
	hash := sha256.New()
	//fmt.Fprintf()
	fmt.Fprintf(hash, "%s%s", params.NodeName, params.ProviderName)
	hashedBytes := hash.Sum(nil)
	dpidBytes := hashedBytes[:4]

	// Convert the bytes to a hexadecimal string
	dpid := hex.EncodeToString(dpidBytes)
	return fmt.Sprintf("br-%s", dpid)
}
