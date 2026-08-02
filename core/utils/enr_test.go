package utils

import (
	"testing"
)

func TestDecodeENRVersion(t *testing.T) {
	// Example ENR from the issue description
	//nolint:lll
	enrStr := "enr:-M64QNyulHqCeOHgo-lIqP_ukjtnbAM1DSMunMRlu8uWYlvuELTnEyQNovl3373XzuqPlxi02zTiSpCID8pWceg1Qq8HhWF6dGVjsTAwLTExMTU1MTExLWVlNmQ0ZTkzLTQxODkzMzcyMDctMmVmZDNmZDYtMjMzOWRhNDWCaWSCdjSCaXCEK4Mr4IlzZWNwMjU2azGhA0A6Sf7KJ1htR_CcPlWalV1WkwtzPipwBkI-jBzY5QWFg3RjcIKd0IN1ZHCCndCDdmVyhjAuODcuMg"

	enrInfo, err := DecodeENR(enrStr)
	if err != nil {
		t.Fatalf("Failed to decode ENR version: %v", err)
	}

	expectedVersion := "0.87.2"
	expectedCCV := "00-11155111-ee6d4e93-4189337207-2efd3fd6-2339da45"

	if enrInfo.Version != expectedVersion {
		t.Errorf("Expected version %s, but got %s", expectedVersion, enrInfo.Version)
	}

	if enrInfo.CompressedComponentsVersion != expectedCCV {
		t.Errorf("Expected CCV %s, but got %s", expectedCCV, enrInfo.CompressedComponentsVersion)
	}
}

func TestDecodeENRVersionEmpty(t *testing.T) {
	enrInfo, err := DecodeENR("")
	if err == nil {
		t.Error("Expected error for empty ENR string")
	}
	if enrInfo != nil {
		t.Errorf("Expected nil ENRInfo for empty ENR, got %+v", enrInfo)
	}
}

func TestDecodeENRVersionWithOnlyVerField(t *testing.T) {
	// TODO: Add test for ENR with only ver field, no aztec field
	// This would be a constructed ENR for testing purposes
	// You'd need a real ENR without aztec field for a proper test
	// For now, we'll test that the function handles missing fields gracefully
}

func TestDecodeENRVersionFromHex(t *testing.T) {
	// The hex version "0x302e38372e32" should decode to "0.87.2"
	hexVersion := "0x302e38372e32"

	result, err := DecodeENRVersionFromHex(hexVersion)
	if err != nil {
		t.Fatalf("Failed to decode hex version: %v", err)
	}

	expected := "0.87.2"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}
