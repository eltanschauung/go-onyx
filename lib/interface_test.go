package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"git.gammaspectra.live/git/go-away/lib/challenge"
)

func TestChallengeAuditEventsHaveQueryableSchema(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	state := &State{}
	request := httptest.NewRequest(
		http.MethodGet,
		"https://example.test/bant/?fragment=md5&__goaway_token=secret",
		nil,
	)
	registration := &challenge.Registration{Name: "proof"}

	state.ChallengeIssued(request, registration, request.URL.String(), logger)
	state.ChallengePassed(request, registration, request.URL.String(), logger)
	state.ChallengeFailed(request, registration, errors.New("synthetic failure"), request.URL.String(), logger)
	state.ChallengeChecked(request, registration, request.URL.String(), logger)

	want := []struct {
		event   string
		status  string
		outcome string
	}{
		{event: "challenge.issued", status: "pending", outcome: "issued"},
		{event: "challenge.passed", status: "passed", outcome: "passed"},
		{event: "challenge.failed", status: "failed", outcome: "challenge_error"},
		{event: "challenge.checked", status: "passed", outcome: "existing_proof"},
	}

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != len(want) {
		t.Fatalf("logged %d lines, want %d", len(lines), len(want))
	}
	for i, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("line %d is not JSON: %v", i, err)
		}
		if entry["event"] != want[i].event || entry["status"] != want[i].status || entry["outcome"] != want[i].outcome {
			t.Fatalf("line %d has unexpected audit fields: %#v", i, entry)
		}
		if redirect, _ := entry["redirect"].(string); strings.Contains(redirect, "secret") || strings.Contains(redirect, "__goaway") {
			t.Fatalf("line %d exposed internal challenge query values: %q", i, redirect)
		}
	}
}

func TestChallengeAuditLoggingIsConcurrentAndLineAtomic(t *testing.T) {
	var output bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&output, nil))
	state := &State{}
	request := httptest.NewRequest(http.MethodGet, "https://example.test/bant/", nil)
	registration := &challenge.Registration{Name: "proof"}

	const count = 128
	var group sync.WaitGroup
	group.Add(count)
	for range count {
		go func() {
			defer group.Done()
			state.ChallengeChecked(request, registration, request.URL.String(), logger)
		}()
	}
	group.Wait()

	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	if len(lines) != count {
		t.Fatalf("logged %d lines, want %d", len(lines), count)
	}
	for i, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("concurrent line %d is not an atomic JSON object: %v", i, err)
		}
		if entry["event"] != "challenge.checked" {
			t.Fatalf("concurrent line %d has event %#v", i, entry["event"])
		}
	}
}
