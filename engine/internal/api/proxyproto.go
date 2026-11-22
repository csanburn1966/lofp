package api

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Reverse-proxy support. In production every connection arrives through nginx,
// so without this all players share nginx's IP and per-IP limits (telnet/SSH
// connection caps, WebSocket caps, auth rate limits) apply to everyone at once.
//
// LOFP_TRUSTED_PROXIES is a comma-separated list of proxy IPs. Connections from
// those IPs must start with a PROXY protocol v1 header (nginx stream:
// `proxy_protocol on;`) on the telnet and SSH ports, and their HTTP requests
// take the client IP from the last X-Forwarded-For entry. Connections from any
// other address are used as-is and their PROXY headers / X-Forwarded-For are
// not trusted.

var (
	trustedProxiesOnce sync.Once
	trustedProxies     map[string]bool
)

func trustedProxySet() map[string]bool {
	trustedProxiesOnce.Do(func() {
		trustedProxies = map[string]bool{}
		for _, p := range strings.Split(os.Getenv("LOFP_TRUSTED_PROXIES"), ",") {
			if ip := net.ParseIP(strings.TrimSpace(p)); ip != nil {
				trustedProxies[ip.String()] = true
			}
		}
	})
	return trustedProxies
}

func isTrustedProxy(addr net.Addr) bool {
	set := trustedProxySet()
	if len(set) == 0 {
		return false
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		host = addr.String()
	}
	ip := net.ParseIP(host)
	return ip != nil && set[ip.String()]
}

// proxiedConn reports the client address from the PROXY header as RemoteAddr.
type proxiedConn struct {
	net.Conn
	remote net.Addr
}

func (c *proxiedConn) RemoteAddr() net.Addr { return c.remote }

// acceptProxyProtocol returns conn unchanged unless it comes from a trusted
// proxy, in which case it reads and strips the PROXY v1 header and returns a
// conn whose RemoteAddr is the real client. Call it from the per-connection
// goroutine: it blocks until the header arrives (5s limit).
func acceptProxyProtocol(conn net.Conn) (net.Conn, error) {
	if !isTrustedProxy(conn.RemoteAddr()) {
		return conn, nil
	}
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	defer conn.SetReadDeadline(time.Time{})

	// Read byte by byte so nothing past the header is consumed. A v1 header is
	// at most 107 bytes including CRLF.
	var line []byte
	buf := make([]byte, 1)
	for len(line) < 107 {
		if _, err := conn.Read(buf); err != nil {
			return nil, fmt.Errorf("reading PROXY header: %w", err)
		}
		line = append(line, buf[0])
		if buf[0] == '\n' {
			break
		}
	}
	remote, err := parseProxyV1(string(line))
	if err != nil {
		return nil, err
	}
	if remote == nil { // PROXY UNKNOWN: keep the proxy's address
		return conn, nil
	}
	return &proxiedConn{Conn: conn, remote: remote}, nil
}

// parseProxyV1 parses "PROXY TCP4 src dst sport dport\r\n". It returns a nil
// address for "PROXY UNKNOWN".
func parseProxyV1(line string) (net.Addr, error) {
	if !strings.HasSuffix(line, "\r\n") {
		return nil, fmt.Errorf("PROXY header not terminated by CRLF")
	}
	f := strings.Fields(strings.TrimSuffix(line, "\r\n"))
	if len(f) < 2 || f[0] != "PROXY" {
		return nil, fmt.Errorf("missing PROXY header")
	}
	if f[1] == "UNKNOWN" {
		return nil, nil
	}
	if (f[1] != "TCP4" && f[1] != "TCP6") || len(f) != 6 {
		return nil, fmt.Errorf("malformed PROXY header")
	}
	ip := net.ParseIP(f[2])
	port, err := strconv.Atoi(f[4])
	if ip == nil || err != nil || port < 0 || port > 65535 {
		return nil, fmt.Errorf("malformed PROXY header addresses")
	}
	return &net.TCPAddr{IP: ip, Port: port}, nil
}

// getClientIP returns the client IP for an HTTP request. Behind a trusted
// proxy it uses the last X-Forwarded-For entry (the one the proxy itself
// appended, which the client cannot forge). With no trusted proxies
// configured, it keeps the legacy behavior of trusting the first entry.
func getClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	fwd := r.Header.Get("X-Forwarded-For")
	if len(trustedProxySet()) > 0 {
		if fwd != "" && isTrustedProxy(&net.TCPAddr{IP: net.ParseIP(host)}) {
			parts := strings.Split(fwd, ",")
			return strings.TrimSpace(parts[len(parts)-1])
		}
		return host
	}
	if ip := r.Header.Get("Fly-Client-IP"); ip != "" {
		return strings.TrimSpace(ip)
	}
	if fwd != "" {
		return strings.TrimSpace(strings.Split(fwd, ",")[0])
	}
	return host
}
