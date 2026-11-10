package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestIsUserKey(t *testing.T) {
	cases := map[string]bool{
		"1f58644b4d9ba951014dae08fed80006":  true,
		"8a5081f479f7ad42017a5bc2c0680010":  true,
		"8A5081F479F7AD42017A5BC2C0680010":  true,
		"k77319":                            false,
		"1f58644b4d9ba951014dae08fed8000":   false,
		"1f58644b4d9ba951014dae08fed800067": false,
		"zf58644b4d9ba951014dae08fed80006":  false,
	}
	for in, want := range cases {
		if got := isUserKey(in); got != want {
			t.Errorf("isUserKey(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestLookupByKey(t *testing.T) {
	key := "1f58644b4d9ba951014dae08fed80006"
	if !lookupByKey(key, false, false) {
		t.Error("expected auto-detected user key")
	}
	if lookupByKey("k77319", false, false) {
		t.Error("expected auto-detected username")
	}
	if !lookupByKey("k77319", true, false) {
		t.Error("expected --key to force key lookup")
	}
	if lookupByKey(key, false, true) {
		t.Error("expected --username to force username lookup")
	}
}

var sampleUsers = []userResult{
	{Query: "k77319", Username: "k77319", UserKey: "1f58644b4d9ba951014dae08fed80006", DisplayName: "Solheim, Bent Andre"},
}

func TestWriteUsers_Text(t *testing.T) {
	var buf bytes.Buffer
	if err := writeUsers(&buf, "text", sampleUsers); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	for _, want := range []string{"k77319", "Solheim, Bent Andre", "1f58644b4d9ba951014dae08fed80006"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in text output, got %q", want, got)
		}
	}
}

func TestWriteUsers_JSON(t *testing.T) {
	var buf bytes.Buffer
	if err := writeUsers(&buf, "json", sampleUsers); err != nil {
		t.Fatal(err)
	}
	var out []userResult
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(out) != 1 || out[0].UserKey != sampleUsers[0].UserKey {
		t.Errorf("unexpected json output: %+v", out)
	}

	buf.Reset()
	if err := writeUsers(&buf, "json", nil); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Errorf("expected empty array for no users, got %q", buf.String())
	}
}

func TestWriteUsers_Markdown(t *testing.T) {
	var buf bytes.Buffer
	if err := writeUsers(&buf, "markdown", sampleUsers); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "| k77319 | Solheim, Bent Andre | 1f58644b4d9ba951014dae08fed80006 |") {
		t.Errorf("unexpected markdown output: %q", buf.String())
	}
}
