package challenge

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.gammaspectra.live/git/go-away/lib/policy"
	"github.com/go-jose/go-jose/v4/jwt"
)

type cookieTestState struct {
	StateInterface
	privateKey            ed25519.PrivateKey
	publicKey             ed25519.PublicKey
	privateKeyFingerprint []byte
}

func newCookieTestState(t *testing.T) *cookieTestState {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	fingerprint := sha256.Sum256(privateKey)
	return &cookieTestState{
		privateKey:            privateKey,
		publicKey:             publicKey,
		privateKeyFingerprint: fingerprint[:],
	}
}

func (s *cookieTestState) PrivateKey() ed25519.PrivateKey { return s.privateKey }
func (s *cookieTestState) PublicKey() ed25519.PublicKey   { return s.publicKey }
func (s *cookieTestState) PrivateKeyFingerprint() []byte  { return s.privateKeyFingerprint }
func (s *cookieTestState) Settings() policy.StateSettings { return policy.StateSettings{} }
func (s *cookieTestState) GetChallenges() Register        { return Register{} }

func newCookieTestRequest(t *testing.T, state StateInterface) (*http.Request, *RequestData) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "https://example.test/bant", nil)
	req.RemoteAddr = "192.0.2.10:12345"
	req, data := CreateRequestData(req, state)
	return req, data
}

func TestVerifyChallengeStateTriesEverySameNameCookie(t *testing.T) {
	state := newCookieTestState(t)
	_, issuer := newCookieTestRequest(t, state)
	issuer.ChallengeMap = TokenChallengeMap{
		"proof": {Ok: true},
	}
	validValue, err := issuer.issueChallengeState(time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}

	req, verifier := newCookieTestRequest(t, state)
	req.AddCookie(&http.Cookie{Name: verifier.cookieName, Value: "stale-invalid-proof"})
	req.AddCookie(&http.Cookie{Name: verifier.cookieName, Value: validValue})

	got, err := verifier.verifyChallengeState()
	if err != nil {
		t.Fatalf("valid cookie after stale duplicate was rejected: %v", err)
	}
	if proof, ok := got["proof"]; !ok || !proof.Ok {
		t.Fatalf("valid proof state was not recovered: %#v", got)
	}
}

func TestStableBrowserKeyIgnoresLanguageChanges(t *testing.T) {
	state := newCookieTestState(t)
	reg := &Registration{Name: "pow", KeyHeaders: StableBrowserKeyHeaders}
	until := time.Now().Add(time.Hour)

	reqEnglish, _ := newCookieTestRequest(t, state)
	reqEnglish.Header.Set("User-Agent", "Example Browser/1.0")
	reqEnglish.Header.Set("Accept-Language", "en-US,en;q=0.9")
	englishKey := GetChallengeKeyForRequest(state, reg, until, reqEnglish)

	reqFrench, _ := newCookieTestRequest(t, state)
	reqFrench.Header.Set("User-Agent", "Example Browser/1.0")
	reqFrench.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")
	frenchKey := GetChallengeKeyForRequest(state, reg, until, reqFrench)

	if !bytes.Equal(englishKey[:], frenchKey[:]) {
		t.Fatal("Accept-Language change invalidated a stable browser proof")
	}

	reqOtherBrowser, _ := newCookieTestRequest(t, state)
	reqOtherBrowser.Header.Set("User-Agent", "Other Browser/2.0")
	otherBrowserKey := GetChallengeKeyForRequest(state, reg, until, reqOtherBrowser)
	if bytes.Equal(englishKey[:], otherBrowserKey[:]) {
		t.Fatal("User-Agent change did not invalidate a browser proof")
	}
}

func TestStableBrowserKeyAcceptsUnexpiredLegacyProof(t *testing.T) {
	state := newCookieTestState(t)
	reg := &Registration{
		Name:             "pow",
		Duration:         time.Hour,
		KeyHeaders:       StableBrowserKeyHeaders,
		LegacyKeyHeaders: MinimalKeyHeaders,
	}
	req, data := newCookieTestRequest(t, state)
	req.Header.Set("User-Agent", "Example Browser/1.0")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	legacyKey := getChallengeKeyForRequest(
		state,
		reg,
		reg.LegacyKeyHeaders,
		data.Expiration(reg.Duration),
		req,
	)
	data.ChallengeMap = TokenChallengeMap{
		reg.Name: {
			Key:    legacyKey[:],
			Ok:     true,
			Expiry: jwt.NumericDate(time.Now().Add(time.Hour).Unix()),
		},
	}

	stableKey := GetChallengeKeyForRequest(state, reg, data.Expiration(reg.Duration), req)
	result, _, err := data.verifyChallenge(reg, stableKey)
	if err != nil {
		t.Fatalf("unexpired legacy proof returned an error: %v", err)
	}
	if !result.Ok() {
		t.Fatalf("unexpired legacy proof was rejected: %v", result)
	}
}

func TestChallengeDiagnosticCategories(t *testing.T) {
	if got := challengeStateErrorCategory(http.ErrNoCookie); got != "missing" {
		t.Fatalf("missing cookie category = %q", got)
	}
	if got := challengeStateErrorCategory(ErrTokenExpired); got != "expired" {
		t.Fatalf("expired cookie category = %q", got)
	}
	if got := challengeStateErrorCategory(errors.New("malformed JWE")); got != "invalid" {
		t.Fatalf("invalid cookie category = %q", got)
	}

	tests := []struct {
		name         string
		tokenPresent bool
		result       VerifyResult
		err          error
		want         string
	}{
		{name: "missing token", result: VerifyResultFail, want: "missing"},
		{name: "valid token", tokenPresent: true, result: VerifyResultOK, want: "valid"},
		{name: "expired token", tokenPresent: true, result: VerifyResultFail, err: ErrTokenExpired, want: "expired"},
		{name: "key mismatch", tokenPresent: true, result: VerifyResultFail, err: ErrVerifyKeyMismatch, want: "key_mismatch"},
		{name: "verification mismatch", tokenPresent: true, result: VerifyResultFail, err: ErrVerifyVerifyMismatch, want: "verification_mismatch"},
		{name: "invalid token", tokenPresent: true, result: VerifyResultFail, err: errors.New("invalid"), want: "invalid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := challengeVerificationCategory(tt.tokenPresent, tt.result, tt.err); got != tt.want {
				t.Fatalf("category = %q, want %q", got, tt.want)
			}
		})
	}
}
