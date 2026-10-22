package http_client

import (
	"context"
	"net"
	"net/http"
	"time"
)

// Shared is reused by every outbound REST/JSON API call (Discord, PayPal,
// Mailtrap, ...) so idle connections are pooled and reused across services
// instead of each maintaining its own separate connection pool.
var Shared = &http.Client{Timeout: 15 * time.Second}
var SharedIpv4 = &http.Client{Timeout: 15 * time.Second, Transport: ipv4Transport()}

func ipv4Transport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	d := &net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}
	t.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		return d.DialContext(ctx, "tcp4", addr)
	}
	return t
}
