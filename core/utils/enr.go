package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/p2p/enr"
	"github.com/ethereum/go-ethereum/rlp"
)

// ENRInfo contains information extracted from an ENR record
type ENRInfo struct {
	Version                     string // From 'ver' field
	CompressedComponentsVersion string // From 'aztec' field (CCV)
}

// DecodeENR extracts the version (ver attribute) and aztec (CompressedComponentsVersion) from an ENR record string
func DecodeENR(enrStr string) (*ENRInfo, error) {
	if enrStr == "" {
		return nil, fmt.Errorf("empty ENR string")
	}

	// Remove "enr:" prefix if present
	enrStr = strings.TrimPrefix(enrStr, "enr:")

	// Decode base64url to bytes
	data, err := base64.RawURLEncoding.DecodeString(enrStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64url: %w", err)
	}

	// Parse RLP data into a record using DecodeRLP
	var record enr.Record
	stream := rlp.NewStream(bytes.NewReader(data), 0)
	if err := record.DecodeRLP(stream); err != nil {
		return nil, fmt.Errorf("failed to decode RLP: %w", err)
	}

	result := &ENRInfo{}

	// Get the "ver" attribute from the ENR
	var version []byte
	if err := record.Load(enr.WithEntry("ver", &version)); err == nil {
		// Convert bytes to string (the version is stored as ASCII bytes)
		result.Version = string(version)
	}

	// Get the "aztec" attribute from the ENR
	var aztec []byte
	if err := record.Load(enr.WithEntry("aztec", &aztec)); err == nil {
		// Convert bytes to string
		result.CompressedComponentsVersion = string(aztec)
	}

	return result, nil
}

// DecodeENRVersionFromHex extracts the version from hex-encoded version data
func DecodeENRVersionFromHex(hexVersion string) (string, error) {
	if hexVersion == "" {
		return "", nil
	}

	// If the hex string starts with "0x", remove it
	if len(hexVersion) >= 2 && hexVersion[:2] == "0x" {
		hexVersion = hexVersion[2:]
	}

	// Decode hex to bytes
	versionBytes, err := hex.DecodeString(hexVersion)
	if err != nil {
		return "", fmt.Errorf("failed to decode version hex: %w", err)
	}

	return string(versionBytes), nil
}
