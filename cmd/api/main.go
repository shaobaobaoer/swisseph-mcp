package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/shaobaobaoer/solarsage-mcp/pkg/api"
	"github.com/shaobaobaoer/solarsage-mcp/pkg/sweph"
)

func main() {
	port := flag.Int("port", 8080, "HTTP port")
	apiKey := flag.String("api-key", "", "API key for authentication (empty = no auth)")
	flag.Parse()

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

	absPath, _ := filepath.Abs(ephePath)
	sweph.Init(absPath)
	sweph.ConfigureFromEnv() // Configure ephemeris type from SWISSEPH_TYPE / SWISSEPH_JPL_FILE
	defer sweph.Close()

	srv := api.NewServer(api.Config{
		APIKey: *apiKey,
		Port:   *port,
	})

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("SolarSage API server starting on %s\n", addr)
	if err := srv.Run(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
