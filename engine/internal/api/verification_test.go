package api

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jonradoff/lofp/internal/auth"
	"github.com/jonradoff/lofp/internal/email"
)

func TestResendVerificationRateLimit(t *testing.T) {
	setTrustedProxies(t, "192.168.4.122")
	s := &Server{
		auth:       auth.NewService(nil, "", strings.Repeat("x", 32), ""),
		rateLimits: make(map[string][]time.Time),
	}
	post := func(clientIP string) int {
		r := httptest.NewRequest("POST", "/api/auth/resend-verification", strings.NewReader(`{"email":"a@example.com"}`))
		r.RemoteAddr = "192.168.4.122:5555"
		r.Header.Set("X-Forwarded-For", clientIP)
		w := httptest.NewRecorder()
		s.handleResendVerification(w, r)
		return w.Code
	}
	// No database, so allowed requests fail with 400.
	for i := 0; i < 3; i++ {
		if got := post("203.0.113.7"); got != 400 {
			t.Fatalf("request %d: got %d, want 400", i+1, got)
		}
	}
	if got := post("203.0.113.7"); got != 429 {
		t.Errorf("4th request: got %d, want 429", got)
	}
	if got := post("203.0.113.8"); got != 400 {
		t.Errorf("other IP: got %d, want 400", got)
	}
}

func TestSendVerificationReportsFailure(t *testing.T) {
	account := &auth.Account{Email: "a@example.com"}
	if err := (&Server{}).sendVerification(account, "t", "c"); err == nil {
		t.Error("nil email service: want error")
	}
	s := &Server{email: email.New("", "x@example.com", "http://localhost")}
	if err := s.sendVerification(account, "t", "c"); err == nil {
		t.Error("unconfigured email service: want error")
	}
}

// Signups share one per-IP limit across web, telnet and SSH.
func TestRegistrationLimitShared(t *testing.T) {
	setTrustedProxies(t, "192.168.4.122")
	s := &Server{
		auth:       auth.NewService(nil, "", strings.Repeat("x", 32), ""),
		rateLimits: make(map[string][]time.Time),
	}

	// Runs telnetRegister over a pipe and returns everything it printed.
	telnetSignup := func(ip string) string {
		server, client := net.Pipe()
		defer client.Close()
		conn := &proxiedConn{Conn: server, remote: &net.TCPAddr{IP: net.ParseIP(ip), Port: 4242}}
		tc := &telnetConn{conn: conn, reader: bufio.NewReader(conn), width: 80}
		go func() {
			client.Write([]byte("Tester\r\na@example.com\r\nGoodPassw0rd!\r\nGoodPassw0rd!\r\n"))
		}()
		go func() {
			s.telnetRegister(tc, context.Background())
			server.Close()
		}()
		out, _ := io.ReadAll(client)
		return string(out)
	}
	web := func(ip string) int {
		r := httptest.NewRequest("POST", "/api/auth/register", strings.NewReader(`{"email":"a@example.com","password":"GoodPassw0rd!","name":"Tester"}`))
		r.RemoteAddr = "192.168.4.122:5555"
		r.Header.Set("X-Forwarded-For", ip)
		w := httptest.NewRecorder()
		s.handleRegister(w, r)
		return w.Code
	}

	const limited = "Too many registration attempts"
	for i := 0; i < 3; i++ {
		if out := telnetSignup("203.0.113.7"); strings.Contains(out, limited) {
			t.Fatalf("telnet signup %d limited too early: %q", i+1, out)
		}
	}
	if out := telnetSignup("203.0.113.7"); !strings.Contains(out, limited) {
		t.Errorf("4th telnet signup not limited: %q", out)
	}
	if got := web("203.0.113.7"); got != 429 {
		t.Errorf("web signup after telnet limit: got %d, want 429", got)
	}
	if got := web("203.0.113.8"); got == 429 {
		t.Error("other IP should not be limited")
	}
}
