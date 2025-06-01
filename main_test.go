package main

import (
	"testing"
)

func TestIsDomainAllowed(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"api.weixin.qq.com", true},
		{"api.wechat.com", true},
		{"mp.weixin.qq.com", true},
		{"qyapi.weixin.qq.com", true},
		{"api.weixin.qq.com:443", true},
		{"sub.api.weixin.qq.com", false},
		{"google.com", false},
		{"evil.api.weixin.qq.com.hacker.com", false},
		{"", false},
	}

	for _, test := range tests {
		result := isDomainAllowed(test.host)
		if result != test.expected {
			t.Errorf("isDomainAllowed(%q) = %v, expected %v", test.host, result, test.expected)
		}
	}
}

func TestVersion(t *testing.T) {
	if Version == "" {
		t.Error("Version should not be empty")
	}
}

func TestAllowedDomains(t *testing.T) {
	expectedDomains := []string{
		"api.weixin.qq.com",
		"api.wechat.com",
		"mp.weixin.qq.com",
		"qyapi.weixin.qq.com",
	}

	if len(allowedDomains) != len(expectedDomains) {
		t.Errorf("Expected %d allowed domains, got %d", len(expectedDomains), len(allowedDomains))
	}

	for i, domain := range expectedDomains {
		if i >= len(allowedDomains) || allowedDomains[i] != domain {
			t.Errorf("Expected domain %q at index %d, got %q", domain, i, allowedDomains[i])
		}
	}
} 