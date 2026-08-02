package core

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"testing"
)

// Helper function to encode a string with length prefix
func encodeString(s string) []byte {
	buf := new(bytes.Buffer)
	length := uint32(len(s)) // #nosec G115
	_ = binary.Write(buf, binary.BigEndian, length)
	buf.WriteString(s)
	return buf.Bytes()
}

// Helper function to encode a uint32
func encodeUint32(n uint32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, n)
	return buf
}

// TestDecodeStatusMessage tests the status message decoder
func TestDecodeStatusMessage(t *testing.T) {
	// Create a test message
	testMsg := StatusMessage{
		CompressedComponentsVersion: "1.0.0-staging.4",
		LatestBlockNumber:           12345,
		LatestBlockHash:             "0xabcdef1234567890",
		FinalisedBlockNumber:        12340,
	}

	// Manually encode the message
	var encoded bytes.Buffer
	encoded.Write(encodeString(testMsg.CompressedComponentsVersion))
	encoded.Write(encodeUint32(testMsg.LatestBlockNumber))
	encoded.Write(encodeString(testMsg.LatestBlockHash))
	encoded.Write(encodeUint32(testMsg.FinalisedBlockNumber))

	// Decode the message
	decoded, err := DecodeStatusMessage(encoded.Bytes())
	if err != nil {
		t.Fatalf("Failed to decode message: %v", err)
	}

	// Verify the decoded message
	if decoded.CompressedComponentsVersion != testMsg.CompressedComponentsVersion {
		t.Errorf("CompressedComponentsVersion mismatch: got %s, want %s",
			decoded.CompressedComponentsVersion, testMsg.CompressedComponentsVersion)
	}
	if decoded.LatestBlockNumber != testMsg.LatestBlockNumber {
		t.Errorf("LatestBlockNumber mismatch: got %d, want %d",
			decoded.LatestBlockNumber, testMsg.LatestBlockNumber)
	}
	if decoded.LatestBlockHash != testMsg.LatestBlockHash {
		t.Errorf("LatestBlockHash mismatch: got %s, want %s",
			decoded.LatestBlockHash, testMsg.LatestBlockHash)
	}
	if decoded.FinalisedBlockNumber != testMsg.FinalisedBlockNumber {
		t.Errorf("FinalisedBlockNumber mismatch: got %d, want %d",
			decoded.FinalisedBlockNumber, testMsg.FinalisedBlockNumber)
	}
}

// TestDecodeStatusMessageErrors tests error handling
func TestDecodeStatusMessageErrors(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "incomplete string length",
			data:    []byte{0, 0, 0}, // Only 3 bytes for string length
			wantErr: true,
		},
		{
			name:    "string length too large",
			data:    []byte{0xFF, 0xFF, 0xFF, 0xFF}, // Very large string length
			wantErr: true,
		},
		{
			name:    "incomplete message",
			data:    append(encodeString("test"), encodeUint32(123)...), // Missing fields
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeStatusMessage(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeStatusMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Example of decoding a base64 encoded message
func ExampleDecodeStatusMessage() {
	// Example: If you have a base64 encoded status message from Aztec P2P
	// First decode the base64
	// base64String := "..." // Your actual base64 string here
	// data, err := base64.StdEncoding.DecodeString(base64String)
	// if err != nil {
	//     log.Fatal(err)
	// }

	// For this example, we'll create a sample message
	var buf bytes.Buffer
	buf.Write(encodeString("1.0.0-staging.4"))
	buf.Write(encodeUint32(42))
	buf.Write(encodeString("0x1234567890abcdef"))
	buf.Write(encodeUint32(40))

	// Decode the message
	msg, err := DecodeStatusMessage(buf.Bytes())
	if err != nil {
		panic(err)
	}

	// Use the decoded message
	println("Version:", msg.CompressedComponentsVersion)
	println("Latest Block:", msg.LatestBlockNumber)
	println("Latest Hash:", msg.LatestBlockHash)
	println("Finalised Block:", msg.FinalisedBlockNumber)
}

// Helper function to demonstrate encoding (for testing purposes)
func EncodeStatusMessage(msg *StatusMessage) []byte {
	var buf bytes.Buffer
	buf.Write(encodeString(msg.CompressedComponentsVersion))
	buf.Write(encodeUint32(msg.LatestBlockNumber))
	buf.Write(encodeString(msg.LatestBlockHash))
	buf.Write(encodeUint32(msg.FinalisedBlockNumber))
	return buf.Bytes()
}

// Example showing how to encode and decode
func TestRoundTrip(t *testing.T) {
	original := &StatusMessage{
		CompressedComponentsVersion: "2.0.0-beta.1",
		LatestBlockNumber:           999999,
		LatestBlockHash:             "0xdeadbeefcafe1234567890abcdef",
		FinalisedBlockNumber:        999990,
	}

	// Encode
	encoded := EncodeStatusMessage(original)

	// Convert to base64 (as it would be transmitted)
	base64Encoded := base64.StdEncoding.EncodeToString(encoded)
	t.Logf("Base64 encoded: %s", base64Encoded)

	// Decode from base64
	decoded64, err := base64.StdEncoding.DecodeString(base64Encoded)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	// Decode the message
	result, err := DecodeStatusMessage(decoded64)
	if err != nil {
		t.Fatalf("Failed to decode message: %v", err)
	}

	// Verify
	if result.CompressedComponentsVersion != original.CompressedComponentsVersion ||
		result.LatestBlockNumber != original.LatestBlockNumber ||
		result.LatestBlockHash != original.LatestBlockHash ||
		result.FinalisedBlockNumber != original.FinalisedBlockNumber {
		t.Errorf("Round trip failed: got %+v, want %+v", result, original)
	}
}

// Test with actual base64 encoded status message from Aztec network
func TestDecodeRealAztecMessage(t *testing.T) {
	// Real base64 encoded status message from Aztec network
	base64Message := "AAAALzAwLTExMTU1MTExLTdjNjcyNDE2LTM2Mjg4MDg1LTAwZDA5ODA2LTFhNTA3OWI1AAAikAAAAEIweDEwODY" +
		"xZmJlMWVkMzg2ODNjMTQ4ZmQwNWU1M2M5ODc0YWZlYjU2ZjYwMTYxYjAyZmFkZDk2MmY5YzgxMjI2ZWMAACIz"

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(base64Message)
	if err != nil {
		t.Fatalf("Failed to decode base64: %v", err)
	}

	// Decode the status message
	msg, err := DecodeStatusMessage(data)
	if err != nil {
		t.Fatalf("Failed to decode status message: %v", err)
	}

	// Log the decoded message
	t.Logf("Decoded real Aztec status message:")
	t.Logf("  CompressedComponentsVersion: %s", msg.CompressedComponentsVersion)
	t.Logf("  LatestBlockNumber: %d", msg.LatestBlockNumber)
	t.Logf("  LatestBlockHash: %s", msg.LatestBlockHash)
	t.Logf("  FinalisedBlockNumber: %d", msg.FinalisedBlockNumber)

	// Verify expected values based on the encoded data
	expectedVersion := "00-11155111-7c672416-36288085-00d09806-1a5079b5"
	expectedLatestBlock := uint32(8848) // 0x2290 in hex
	expectedBlockHash := "0x10861fbe1ed38683c148fd05e53c9874afeb56f60161b02fadd962f9c81226ec"
	expectedFinalizedBlock := uint32(8755) // 0x2233 in hex

	if msg.CompressedComponentsVersion != expectedVersion {
		t.Errorf("Version mismatch: got %s, want %s", msg.CompressedComponentsVersion, expectedVersion)
	}
	if msg.LatestBlockNumber != expectedLatestBlock {
		t.Errorf("Latest block mismatch: got %d, want %d", msg.LatestBlockNumber, expectedLatestBlock)
	}
	if msg.LatestBlockHash != expectedBlockHash {
		t.Errorf("Block hash mismatch: got %s, want %s", msg.LatestBlockHash, expectedBlockHash)
	}
	if msg.FinalisedBlockNumber != expectedFinalizedBlock {
		t.Errorf("Finalised block mismatch: got %d, want %d", msg.FinalisedBlockNumber, expectedFinalizedBlock)
	}
}
