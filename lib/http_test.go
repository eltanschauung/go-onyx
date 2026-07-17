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

func TestChallengeStateClearRequiresSameOriginAjax(t *testing.T) {
	tests := []struct {
		name      string
		requested string
		site      string
		want      bool
	}{
		{name: "same origin", requested: "XMLHttpRequest", site: "same-origin", want: true},
		{name: "legacy same origin", requested: "XMLHttpRequest", want: true},
		{name: "missing ajax header", site: "same-origin", want: false},
		{name: "cross site", requested: "XMLHttpRequest", site: "cross-site", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "https://example.test/.go-away/clear-state", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Requested-With", tt.requested)
			req.Header.Set("Sec-Fetch-Site", tt.site)
			if got := allowChallengeStateClear(req); got != tt.want {
				t.Fatalf("allowChallengeStateClear() = %v, want %v", got, tt.want)
			}
		})
	}
}
