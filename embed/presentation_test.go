package embed

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image/png"
	"strings"
	"testing"
)

func TestOnyxAnubisTemplate(t *testing.T) {
	data, err := TemplatesFs.ReadFile("challenge-anubis.gohtml")
	if err != nil {
		t.Fatal(err)
	}
	template := string(data)

	if !strings.Contains(template, "max-width:384px") {
		t.Fatal("Onyx logo is not allowed to render at its native width")
	}
	for _, unwanted := range []string{
		"details_contact_admin_with_request_id",
		"<footer>",
		"Protected by",
	} {
		if strings.Contains(template, unwanted) {
			t.Fatalf("template still contains removed content %q", unwanted)
		}
	}
}

func TestOnyxLogo(t *testing.T) {
	data, err := AssetsFs.ReadFile("static/logo.png")
	if err != nil {
		t.Fatal(err)
	}

	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	if config.Width != 384 || config.Height != 384 {
		t.Fatalf("unexpected logo dimensions %dx%d", config.Width, config.Height)
	}

	want := "751c77c1d0de51bb23b82a3b3f2fdb6069d870d251ca0853113ca072677e8f2f"
	got := sha256.Sum256(data)
	if hex.EncodeToString(got[:]) != want {
		t.Fatalf("unexpected logo digest %x", got)
	}
}
