package lib

import (
	"net/http"
	"testing"
)

func TestCleanupProxyRequestPreservesNormalReferer(t *testing.T) {
	req, err := http.NewRequest("GET", "https://example.test/bant?fragment=md5", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Referer", "https://example.test/bant/res/388371.html")

	cleanupProxyRequest(req)

	if got := req.Referer(); got != "https://example.test/bant/res/388371.html" {
		t.Fatalf("normal Referer changed to %q", got)
	}
}

func TestCleanupProxyRequestRestoresSavedReferer(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		"https://example.test/bant?fragment=md5&__goaway_referer=https%3A%2F%2Fexample.test%2Fbant%2Fres%2F388371.html",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Referer", "https://example.test/.go-away/challenge/verify-challenge")

	cleanupProxyRequest(req)

	if got := req.Referer(); got != "https://example.test/bant/res/388371.html" {
		t.Fatalf("saved Referer was not restored: %q", got)
	}
	if got := req.URL.RawQuery; got != "fragment=md5" {
		t.Fatalf("internal challenge parameters leaked upstream: %q", got)
	}
}

func TestCleanupProxyRequestPreservesLegacyBareQuery(t *testing.T) {
	req, err := http.NewRequest(
		"GET",
		"https://example.test/mod.php?/bant/delete/388371/signed-token&__goaway_challenge=js-pow-sha256",
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	cleanupProxyRequest(req)

	if got := req.URL.RawQuery; got != "/bant/delete/388371/signed-token" {
		t.Fatalf("legacy signed query changed to %q", got)
	}
}
