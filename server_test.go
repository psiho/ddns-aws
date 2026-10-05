package main

import (
	"net"
	"net/http/httptest"
	"testing"
)

func TestIsPublicIPv4(t *testing.T) {
	cases := map[string]bool{
		"109.60.12.42":  true,
		"88.207.75.178": true,
		"192.168.0.100": false,
		"10.1.2.3":      false,
		"172.17.0.2":    false,
		"100.64.1.1":    false,
		"100.128.0.1":   true,
		"127.0.0.1":     false,
		"0.0.0.0":       false,
		"169.254.1.1":   false,
		"2a00:1450::1":  false,
	}
	for in, want := range cases {
		if got := isPublicIPv4(net.ParseIP(in)); got != want {
			t.Errorf("isPublicIPv4(%v) = %v, want %v", in, got, want)
		}
	}
}

func TestClientIP(t *testing.T) {
	cases := []struct {
		remote, realIP, want string
	}{
		{"127.0.0.1:5555", "109.60.12.42", "109.60.12.42"},  // local nginx
		{"172.17.0.1:5555", "109.60.12.42", "109.60.12.42"}, // docker nginx
		{"127.0.0.1:5555", "", "127.0.0.1"},                 // proxy without header
		{"203.0.113.9:5555", "109.60.12.42", "203.0.113.9"}, // direct client, header ignored
		{"[::1]:5555", "109.60.12.42", "109.60.12.42"},      // local nginx over IPv6
	}
	for _, c := range cases {
		r := httptest.NewRequest("GET", "/nic/update", nil)
		r.RemoteAddr = c.remote
		if c.realIP != "" {
			r.Header.Set("X-Real-IP", c.realIP)
		}
		got, err := clientIP(r)
		if err != nil || got.String() != c.want {
			t.Errorf("clientIP(%v, X-Real-IP=%v) = %v, %v; want %v", c.remote, c.realIP, got, err, c.want)
		}
	}
}
