package cmd

import (
	"testing"
)

func TestBuildTextSearchCQL_NoSpaces(t *testing.T) {
	got := buildTextSearchCQL("GitHub-pilotering", "")
	want := `siteSearch ~ "GitHub-pilotering" AND type = page`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildTextSearchCQL_SingleSpace(t *testing.T) {
	got := buildTextSearchCQL("deployment pipeline", "MUP")
	if got != `siteSearch ~ "deployment pipeline" AND type = page AND space in ("MUP")` {
		t.Errorf("got %q", got)
	}
}

func TestBuildTextSearchCQL_MultipleSpaces(t *testing.T) {
	got := buildTextSearchCQL("test", "MUP,DEV,OPS")
	if got != `siteSearch ~ "test" AND type = page AND space in ("MUP","DEV","OPS")` {
		t.Errorf("got %q", got)
	}
}

func TestBuildTextSearchCQL_SpacesWithWhitespace(t *testing.T) {
	got := buildTextSearchCQL("test", " MUP , DEV , ")
	if got != `siteSearch ~ "test" AND type = page AND space in ("MUP","DEV")` {
		t.Errorf("got %q", got)
	}
}

func TestBuildTextSearchCQL_OnlyWhitespaceSpaces(t *testing.T) {
	got := buildTextSearchCQL("test", "   ")
	want := `siteSearch ~ "test" AND type = page`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseSpaces_Empty(t *testing.T) {
	if got := parseSpaces(""); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestParseSpaces_Normal(t *testing.T) {
	got := parseSpaces("A,B,C")
	if len(got) != 3 || got[0] != "A" || got[1] != "B" || got[2] != "C" {
		t.Errorf("expected [A B C], got %v", got)
	}
}

func TestParseSpaces_TrimsAndSkipsEmpty(t *testing.T) {
	got := parseSpaces(" A , , B ")
	if len(got) != 2 || got[0] != "A" || got[1] != "B" {
		t.Errorf("expected [A B], got %v", got)
	}
}
