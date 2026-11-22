package api

import (
	"io"
	"net"
	"net/http/httptest"
	"sync"
	"testing"
)

func setTrustedProxies(t *testing.T, v string) {
	t.Helper()
	t.Setenv("LOFP_TRUSTED_PROXIES", v)
	trustedProxiesOnce = sync.Once{}
	t.Cleanup(func() { trustedProxiesOnce = sync.Once{} })
}

func TestParseProxyV1(t *testing.T) {
	cases := []struct {
		line, want string
		wantErr    bool
	}{
		{"PROXY TCP4 203.0.113.7 192.168.4.36 51234 4000\r\n", "203.0.113.7:51234", false},
		{"PROXY TCP6 2001:db8::1 2001:db8::2 51234 4000\r\n", "[2001:db8::1]:51234", false},
		{"PROXY UNKNOWN\r\n", "", false},
		{"PROXY TCP4 203.0.113.7 192.168.4.36 51234 4000\n", "", true},
		{"SSH-2.0-OpenSSH_9.6\r\n", "", true},
		{"PROXY TCP4 notanip 192.168.4.36 51234 4000\r\n", "", true},
		{"PROXY TCP4 203.0.113.7 192.168.4.36 99999 4000\r\n", "", true},
	}
	for _, c := range cases {
		addr, err := parseProxyV1(c.line)
		if (err != nil) != c.wantErr {
			t.Errorf("%q: err=%v, wantErr=%v", c.line, err, c.wantErr)
			continue
		}
		got := ""
		if addr != nil {
			got = addr.String()
		}
		if !c.wantErr && got != c.want {
			t.Errorf("%q: got %q, want %q", c.line, got, c.want)
		}
	}
}

// dial opens a loopback TCP connection, writes payload, and returns the
// server side of it.
func dial(t *testing.T, payload string) net.Conn {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	client, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { client.Close() })
	if _, err := client.Write([]byte(payload)); err != nil {
		t.Fatal(err)
	}
	server, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { server.Close() })
	return server
}

func TestAcceptProxyProtocolTrusted(t *testing.T) {
	setTrustedProxies(t, "127.0.0.1")
	conn, err := acceptProxyProtocol(dial(t, "PROXY TCP4 203.0.113.7 127.0.0.1 51234 4000\r\nhello"))
	if err != nil {
		t.Fatal(err)
	}
	if got := conn.RemoteAddr().String(); got != "203.0.113.7:51234" {
		t.Errorf("RemoteAddr = %s", got)
	}
	// Bytes after the header must still be readable.
	buf := make([]byte, 5)
	if _, err := io.ReadFull(conn, buf); err != nil || string(buf) != "hello" {
		t.Errorf("payload = %q, err = %v", buf, err)
	}
}

func TestAcceptProxyProtocolTrustedWithoutHeader(t *testing.T) {
	setTrustedProxies(t, "127.0.0.1")
	if _, err := acceptProxyProtocol(dial(t, "SSH-2.0-OpenSSH_9.6\r\n")); err == nil {
		t.Error("expected a trusted proxy connection without a PROXY header to be rejected")
	}
}

func TestAcceptProxyProtocolUntrusted(t *testing.T) {
	setTrustedProxies(t, "192.0.2.1")
	raw := dial(t, "PROXY TCP4 203.0.113.7 127.0.0.1 51234 4000\r\n")
	conn, err := acceptProxyProtocol(raw)
	if err != nil {
		t.Fatal(err)
	}
	if conn != raw {
		t.Error("untrusted connection must be returned unchanged (header not honored)")
	}
}

func TestGetClientIP(t *testing.T) {
	req := func(remote, xff string) string {
		r := httptest.NewRequest("GET", "/", nil)
		r.RemoteAddr = remote
		if xff != "" {
			r.Header.Set("X-Forwarded-For", xff)
		}
		return getClientIP(r)
	}

	setTrustedProxies(t, "192.168.4.122")
	if got := req("192.168.4.122:5555", "6.6.6.6, 203.0.113.7"); got != "203.0.113.7" {
		t.Errorf("trusted proxy: got %s, want last XFF entry", got)
	}
	if got := req("198.51.100.9:5555", "6.6.6.6"); got != "198.51.100.9" {
		t.Errorf("untrusted peer: got %s, want RemoteAddr (XFF ignored)", got)
	}
	if got := req("192.168.4.122:5555", ""); got != "192.168.4.122" {
		t.Errorf("trusted proxy without XFF: got %s", got)
	}

	setTrustedProxies(t, "")
	if got := req("198.51.100.9:5555", "6.6.6.6, 203.0.113.7"); got != "6.6.6.6" {
		t.Errorf("legacy mode: got %s, want first XFF entry", got)
	}
}
