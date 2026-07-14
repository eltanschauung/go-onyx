package challenge

import (
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"git.gammaspectra.live/git/go-away/lib/policy"
)

type cookieTestState struct {
	StateInterface
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
}

func newCookieTestState(t *testing.T) *cookieTestState {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return &cookieTestState{privateKey: privateKey, publicKey: publicKey}
}

func (s *cookieTestState) PrivateKey() ed25519.PrivateKey { return s.privateKey }
func (s *cookieTestState) PublicKey() ed25519.PublicKey   { return s.publicKey }
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
