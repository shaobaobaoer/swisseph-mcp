package main

import (
	"fmt"
	"os"
	"path/filepath"

	swisseph "github.com/anthropic/swisseph-mcp"
	"github.com/anthropic/swisseph-mcp/pkg/mcp"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("swisseph-mcp %s\n", swisseph.Version)
			return
		case "--help", "-h":
			fmt.Println("swisseph-mcp - Astrology MCP server powered by Swiss Ephemeris")
			fmt.Printf("Version: %s\n\n", swisseph.Version)
			fmt.Println("Usage: swisseph-mcp [options]")
			fmt.Println()
			fmt.Println("Options:")
			fmt.Println("  --version, -v  Print version")
			fmt.Println("  --help, -h     Print this help")
			fmt.Println()
			fmt.Println("Environment:")
			fmt.Println("  SWISSEPH_EPHE_PATH  Path to Swiss Ephemeris data files")
			return
		}
	}

	ephePath := os.Getenv("SWISSEPH_EPHE_PATH")
	if ephePath == "" {
		exe, err := os.Executable()
		if err == nil {
			ephePath = filepath.Join(filepath.Dir(exe), "..", "..", "third_party", "swisseph", "ephe")
		}
		if _, err := os.Stat(ephePath); err != nil {
			ephePath = filepath.Join(".", "third_party", "swisseph", "ephe")
		}
	}

	if _, err := os.Stat(ephePath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: ephemeris path not found: %s\n", ephePath)
		fmt.Fprintf(os.Stderr, "Set SWISSEPH_EPHE_PATH environment variable\n")
	}

	server := mcp.NewServer(ephePath)
	if err := server.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
