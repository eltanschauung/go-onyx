package challenge

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestVerifiedImmutableMediaUsesPrivateBrowserCache(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://example.test/bant/thumb/example.png", nil)
	data := verifiedResponseData()
	w := NewResponsePolicyWriter(recorder, request, data)

	data.applyBroadVary(w)
	w.Header().Add("Vary", "Accept-Encoding")
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, s-maxage=600, immutable")
	w.WriteHeader(http.StatusOK)

	if got := recorder.Header().Get("Cache-Control"); got != "private, max-age=31536000, immutable" {
		t.Fatalf("unexpected Cache-Control: %q", got)
	}
	if got := recorder.Header().Values("Vary"); !slices.Equal(got, []string{"Accept-Encoding"}) {
		t.Fatalf("unexpected Vary values: %q", got)
	}
}

func TestVerifiedImmutableMediaRemovesOnlyGoAwayVaryValue(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://example.test/bant/thumb/example.png", nil)
	data := verifiedResponseData()
	w := NewResponsePolicyWriter(recorder, request, data)

	data.applyBroadVary(w)
	w.Header().Add("Vary", broadVaryHeaderValue)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)

	if got := recorder.Header().Values("Vary"); !slices.Equal(got, []string{broadVaryHeaderValue}) {
		t.Fatalf("upstream Vary value was not preserved: %q", got)
	}
}

func TestImmutableMediaPolicyRequiresVerification(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://example.test/bant/thumb/example.png", nil)
	data := &RequestData{ChallengeVerify: map[Id]VerifyResult{}}
	w := NewResponsePolicyWriter(recorder, request, data)

	data.applyBroadVary(w)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)

	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("unverified response Cache-Control changed to %q", got)
	}
	if got := recorder.Header().Get("Vary"); got != broadVaryHeaderValue {
		t.Fatalf("unverified response Vary changed to %q", got)
	}
}

func TestVerifiedMutableMediaKeepsBroadVary(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://example.test/bant/src/example.png", nil)
	data := verifiedResponseData()
	w := NewResponsePolicyWriter(recorder, request, data)

	data.applyBroadVary(w)
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=2592000")
	w.WriteHeader(http.StatusOK)

	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=2592000" {
		t.Fatalf("mutable response Cache-Control changed to %q", got)
	}
	if got := recorder.Header().Get("Vary"); got != broadVaryHeaderValue {
		t.Fatalf("mutable response Vary changed to %q", got)
	}
}

func TestVerifiedImmutableNonMediaKeepsBroadVary(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "https://example.test/assets/application.js", nil)
	data := verifiedResponseData()
	w := NewResponsePolicyWriter(recorder, request, data)

	data.applyBroadVary(w)
	w.Header().Set("Content-Type", "application/javascript")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.WriteHeader(http.StatusOK)

	if got := recorder.Header().Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
		t.Fatalf("non-media response Cache-Control changed to %q", got)
	}
	if got := recorder.Header().Get("Vary"); got != broadVaryHeaderValue {
		t.Fatalf("non-media response Vary changed to %q", got)
	}
}

func TestVerifiedImmutableMediaPathSupportsHeaderlessNotModified(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodHead, "https://example.test/bant/thumb/example.webm", nil)
	data := verifiedResponseData()
	w := NewResponsePolicyWriter(recorder, request, data)

	data.applyBroadVary(w)
	w.Header().Set("Cache-Control", "PUBLIC, MAX-AGE=31536000, IMMUTABLE")
	w.WriteHeader(http.StatusNotModified)

	if got := recorder.Header().Get("Cache-Control"); got != "private, MAX-AGE=31536000, IMMUTABLE" {
		t.Fatalf("unexpected Cache-Control: %q", got)
	}
	if got := recorder.Header().Values("Vary"); len(got) != 0 {
		t.Fatalf("Go-Onyx Vary was not removed: %q", got)
	}
}

func verifiedResponseData() *RequestData {
	return &RequestData{
		ChallengeVerify: map[Id]VerifyResult{
			Id(1): VerifyResultOK,
		},
	}
}
