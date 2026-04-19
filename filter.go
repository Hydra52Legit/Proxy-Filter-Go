package main

import (
	"fmt"
	"net/http"
	"strings"
)

// Filter evaluates HTTP requests against a blacklist
type Filter struct {
	config *Config
}

// NewFilter creates a new filter instance with the given configuration
func NewFilter(config *Config) *Filter {
	if config == nil {
		panic("filter cannot be created with nil config")
	}
	return &Filter{config: config}
}

// CheckRequest determines if a request should be blocked
func (f *Filter) CheckRequest(req *http.Request) (blocked bool, reason string) {
	host := extractHost(req)
	
	if f.config.IsBlocked(host) {
		return true, fmt.Sprintf("domain '%s' is blacklisted", host)
	}
	
	if parent := f.checkSubdomains(host); parent != "" {
		return true, fmt.Sprintf("subdomain '%s' (parent: '%s') is blacklisted", host, parent)
	}
	
	return false, ""
}

// extractHost gets the hostname from request, removing port if present
func extractHost(req *http.Request) string {
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	
	// Remove port number if exists
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	
	return host
}

// checkSubdomains checks if any parent domain is blacklisted
func (f *Filter) checkSubdomains(host string) string {
	parts := strings.Split(host, ".")
	
	// Start from second-level domain upwards
	// Example: "sub.ads.google.com" -> checks "ads.google.com", then "google.com"
	for i := 1; i < len(parts); i++ {
		domain := strings.Join(parts[i:], ".")
		if f.config.IsBlocked(domain) {
			return domain
		}
	}
	
	return ""
}