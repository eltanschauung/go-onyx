package settings

import (
	"strings"
	"testing"
)

func TestOnyxChallengeStrings(t *testing.T) {
	if got := DefaultStrings.Get("title_challenge"); got != "Confirming you are a whale" {
		t.Fatalf("unexpected challenge title %q", got)
	}

	if got := DefaultStrings.Get("status_calculating"); got != "Onyx is pouring sproke into your computer..." {
		t.Fatalf("unexpected calculation status %q", got)
	}

	details := string(DefaultStrings.Get("details_text"))
	if !strings.Contains(details, `href="/auth"`) || !strings.Contains(details, "bantculture.com/auth") {
		t.Fatalf("challenge explanation does not link to Bantculture authentication: %q", details)
	}
}
