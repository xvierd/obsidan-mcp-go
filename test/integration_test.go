package main

import (
	"fmt"
	"os"
	"testing"
)

// TestBinaryExists verifies the binary was built
func TestBinaryExists(t *testing.T) {
	binaryPath := "../build/mcp-obsidian-go"
	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatalf("Binary not found at %s: %v", binaryPath, err)
	}

	if info.Size() == 0 {
		t.Error("Binary exists but is empty")
	}

	// Check size is reasonable (between 1MB and 50MB)
	if info.Size() < 1024*1024 {
		t.Errorf("Binary too small: %d bytes", info.Size())
	}
	if info.Size() > 50*1024*1024 {
		t.Errorf("Binary too large: %d bytes", info.Size())
	}

	fmt.Printf("Server binary size: %.2f MB\n", float64(info.Size())/(1024*1024))
}

// TestBinaryIsExecutable verifies the server binary is executable
func TestBinaryIsExecutable(t *testing.T) {
	serverPath := "../build/mcp-obsidian-go"

	info, err := os.Stat(serverPath)
	if err != nil {
		t.Fatalf("Cannot stat server binary: %v", err)
	}

	if info.Mode()&0111 == 0 {
		t.Error("Server binary is not executable")
	}

	fmt.Printf("Server binary is executable\n")
}
