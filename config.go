// config.go - configuration loading and management for the proxy server
// Memory optimization: using struct{} instead of bool (0 bytes vs 1 byte per entry)
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)
// Constants instead of magic numbers
const (
    DefaultPort       = 8080
    DefaultListenAddr = "127.0.0.1:8080"
)
// Config holds the server configuration and blacklist
type Config struct {
	Port       int
	ListenAddr string
	Blacklist  map[string]struct{} // struct{} consumes 0 bytes of memory
}

// LoadConfig reads blacklist from file and returns a Config struct
func LoadConfig(blacklistFile string) (*Config, error) {
	
	config := &Config{
		Port:       DefaultPort,
		ListenAddr: DefaultListenAddr,
		Blacklist:  make(map[string]struct{}),
	}

	file, err := os.Open(blacklistFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️ Blacklist file %s not found: %v\n", blacklistFile, err)
		fmt.Fprintln(os.Stderr, "📝 Continuing with empty blacklist")
		return config, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Add domain to blacklist (case-insensitive)
		domain := strings.ToLower(line)
		config.Blacklist[domain] = struct{}{}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading blacklist file %s: %w", blacklistFile, err)
	}

	fmt.Printf("✅ Loaded %d domains into blacklist\n", len(config.Blacklist))
	return config, nil
}

// IsBlocked checks if a domain is in the blacklist
func (c *Config) IsBlocked(domain string) bool {
	// Guard against nil receiver (prevents panic)
	if c == nil || c.Blacklist == nil {
		return false
	}
	_, exists := c.Blacklist[strings.ToLower(domain)]
	return exists
}
