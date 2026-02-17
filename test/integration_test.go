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
	
	fmt.Printf("✓ Server binary size: %.2f MB\n", float64(info.Size())/(1024*1024))
}

// TestIndexerBinaryExists verifies the indexer binary was built
func TestIndexerBinaryExists(t *testing.T) {
	binaryPath := "../build/mcp-obsidian-indexer"
	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatalf("Indexer binary not found at %s: %v", binaryPath, err)
	}
	
	if info.Size() == 0 {
		t.Error("Indexer binary exists but is empty")
	}
	
	fmt.Printf("✓ Indexer binary size: %.2f MB\n", float64(info.Size())/(1024*1024))
}

// TestBinaryIsExecutable verifies binaries are executable
func TestBinaryIsExecutable(t *testing.T) {
	serverPath := "../build/mcp-obsidian-go"
	indexerPath := "../build/mcp-obsidian-indexer"
	
	// Check server is executable
	info, err := os.Stat(serverPath)
	if err != nil {
		t.Fatalf("Cannot stat server binary: %v", err)
	}
	
	if info.Mode() & 0111 == 0 {
		t.Error("Server binary is not executable")
	}
	
	// Check indexer is executable
	info, err = os.Stat(indexerPath)
	if err != nil {
		t.Fatalf("Cannot stat indexer binary: %v", err)
	}
	
	if info.Mode() & 0111 == 0 {
		t.Error("Indexer binary is not executable")
	}
	
	fmt.Printf("✓ Both binaries are executable\n")
}
