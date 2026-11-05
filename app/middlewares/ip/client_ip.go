package ip_middleware

import (
	"net"
	"net/http"
	"strings"
)

// ResolveClientIP uses proxy headers only when the direct peer is a trusted
// proxy. Without a configured trusted proxy list, headers are ignored and the
// connection address is kept.
func ResolveClientIP(trustedProxies string) func(http.Handler) http.Handler {
	trusted := parseTrustedProxies(trustedProxies)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(trusted) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			peer := net.ParseIP(strings.TrimSpace(host))
			if peer == nil || !isTrusted(peer, trusted) {
				next.ServeHTTP(w, r)
				return
			}

			client := forwardedIP(r)
			if client != nil {
				r.RemoteAddr = client.String()
			}
			next.ServeHTTP(w, r)
		})
	}
}

func forwardedIP(r *http.Request) net.IP {
	if value := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); value != "" {
		if ip := net.ParseIP(value); ip != nil {
			return ip
		}
	}

	value := strings.TrimSpace(r.Header.Get("X-Forwarded-For"))
	if value == "" {
		return nil
	}
	first := strings.TrimSpace(strings.Split(value, ",")[0])
	return net.ParseIP(first)
}

func parseTrustedProxies(raw string) []*net.IPNet {
	var networks []*net.IPNet
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if _, network, err := net.ParseCIDR(entry); err == nil {
			networks = append(networks, network)
			continue
		}
		if ip := net.ParseIP(entry); ip != nil {
			bits := 128
			if ip.To4() != nil {
				bits = 32
			}
			networks = append(networks, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
		}
	}
	return networks
}

func isTrusted(ip net.IP, networks []*net.IPNet) bool {
	for _, network := range networks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}
