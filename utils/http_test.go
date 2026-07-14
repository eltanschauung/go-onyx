package utils

import "testing"

func TestRawQueryRoundTripPreservesBareComponents(t *testing.T) {
	tests := map[string]string{
		"/bant/delete/388371/signed-token":                      "/bant/delete/388371/signed-token",
		"/reports/12/dismiss&all/signed-token":                  "/reports/12/dismiss&all/signed-token",
		"z=last&flag&a=&dup=one&dup=two&plus=a+b&encoded=a%2Fb": "a&dup=one&dup=two&encoded=a%2Fb&flag&plus=a+b&z=last",
	}

	for raw, expected := range tests {
		t.Run(raw, func(t *testing.T) {
			values, err := ParseRawQuery(raw)
			if err != nil {
				t.Fatalf("ParseRawQuery(%q): %v", raw, err)
			}
			if got := EncodeRawQuery(values); got != expected {
				t.Fatalf("EncodeRawQuery(ParseRawQuery(%q)) = %q, want %q", raw, got, expected)
			}
		})
	}
}
