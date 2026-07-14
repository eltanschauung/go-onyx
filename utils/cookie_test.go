package utils

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestSetCookieUsesSecureRootScope(t *testing.T) {
	req := httptest.NewRequest("GET", "https://example.test/board/thread", nil)
	recorder := httptest.NewRecorder()
	expiry := time.Now().Add(time.Hour)

	SetCookie("challenge", "proof", expiry, recorder, req)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if !cookie.Secure || !cookie.HttpOnly {
		t.Fatalf("challenge cookie must be Secure and HttpOnly: %#v", cookie)
	}
	if cookie.Path != "/" {
		t.Fatalf("expected root cookie path, got %q", cookie.Path)
	}
	if cookie.Domain != "example.test" {
		t.Fatalf("expected request host as cookie domain, got %q", cookie.Domain)
	}
}

func TestClearCookieMatchesCreationScope(t *testing.T) {
	req := httptest.NewRequest("GET", "https://example.test/board/thread", nil)
	recorder := httptest.NewRecorder()

	ClearCookie("challenge", recorder, req)

	cookies := recorder.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("expected one cookie, got %d", len(cookies))
	}

	cookie := cookies[0]
	if cookie.MaxAge != -1 {
		t.Fatalf("expected deletion MaxAge, got %d", cookie.MaxAge)
	}
	if !cookie.Secure || !cookie.HttpOnly {
		t.Fatalf("cleared cookie must retain Secure and HttpOnly: %#v", cookie)
	}
	if cookie.Path != "/" || cookie.Domain != "example.test" {
		t.Fatalf("clear scope does not match creation scope: %#v", cookie)
	}
}
