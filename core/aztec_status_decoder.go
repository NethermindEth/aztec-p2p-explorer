package core

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// StatusMessage represents the P2P status message from Aztec node
type StatusMessage struct {
	CompressedComponentsVersion string
	LatestBlockNumber           uint32
	LatestBlockHash             string
	FinalisedBlockNumber        uint32
}

// BufferReader provides methods to read serialised data
type BufferReader struct {
	data   []byte
	offset int
}

// NewBufferReader creates a new BufferReader from byte slice
func NewBufferReader(data []byte) *BufferReader {
	return &BufferReader{
		data:   data,
		offset: 0,
	}
}

// ReadUint32 reads a 4-byte unsigned integer in big-endian format
func (r *BufferReader) ReadUint32() (uint32, error) {
	if r.offset+4 > len(r.data) {
		return 0, io.ErrUnexpectedEOF
	}
	value := binary.BigEndian.Uint32(r.data[r.offset : r.offset+4])
	r.offset += 4
	return value, nil
}

// ReadString reads a length-prefixed string
func (r *BufferReader) ReadString() (string, error) {
	// Read the 4-byte length prefix
	length, err := r.ReadUint32()
	if err != nil {
		return "", err
	}

	// Check if we have enough bytes for the string
	if r.offset+int(length) > len(r.data) {
		return "", io.ErrUnexpectedEOF
	}

	// Read the string bytes
	str := string(r.data[r.offset : r.offset+int(length)])
	r.offset += int(length)
	return str, nil
}

// DecodeStatusMessage decodes a StatusMessage from binary data
func DecodeStatusMessage(data []byte) (*StatusMessage, error) {
	reader := NewBufferReader(data)
	msg := &StatusMessage{}

	// Read compressedComponentsVersion (string)
	var err error
	msg.CompressedComponentsVersion, err = reader.ReadString()
	if err != nil {
		return nil, fmt.Errorf("failed to read compressedComponentsVersion: %w", err)
	}

	// Read latestBlockNumber (uint32)
	msg.LatestBlockNumber, err = reader.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to read latestBlockNumber: %w", err)
	}

	// Read latestBlockHash (string)
	msg.LatestBlockHash, err = reader.ReadString()
	if err != nil {
		return nil, fmt.Errorf("failed to read latestBlockHash: %w", err)
	}

	// Read finalisedBlockNumber (uint32)
	msg.FinalisedBlockNumber, err = reader.ReadUint32()
	if err != nil {
		return nil, fmt.Errorf("failed to read finalisedBlockNumber: %w", err)
	}

	// Verify all data was consumed
	if reader.offset != len(reader.data) {
		return nil, errors.New("extra data after status message")
	}

	return msg, nil
}

// String returns a string representation of the StatusMessage
func (msg *StatusMessage) String() string {
	return fmt.Sprintf("StatusMessage{\n"+
		"  CompressedComponentsVersion: %s\n"+
		"  LatestBlockNumber: %d\n"+
		"  LatestBlockHash: %s\n"+
		"  FinalisedBlockNumber: %d\n"+
		"}", msg.CompressedComponentsVersion, msg.LatestBlockNumber,
		msg.LatestBlockHash, msg.FinalisedBlockNumber)
}
