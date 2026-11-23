package api

import (
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
