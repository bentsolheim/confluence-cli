package cmd

import (
	"testing"
)

func TestParsePageInput_PlainID(t *testing.T) {
	ref, err := parsePageInput("12345")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ID != "12345" {
		t.Errorf("expected ID=12345, got %q", ref.ID)
	}
}

func TestParsePageInput_SpacesURL(t *testing.T) {
	input := "https://wiki.sits.no/spaces/~k77319/pages/1481704261/Page+Title"
	ref, err := parsePageInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ID != "1481704261" {
		t.Errorf("expected ID=1481704261, got %q", ref.ID)
	}
}

func TestParsePageInput_SpacesURL_URLEncoded(t *testing.T) {
	input := "https://wiki.sits.no/spaces/MUP/pages/1460822101/2026-02-06+Feil+i+funksjonalitet+for+synkronisering+av+delte+hemmeligheter+i+Selvbetjeningsl%C3%B8sning+p%C3%A5+GitHub"
	ref, err := parsePageInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ID != "1460822101" {
		t.Errorf("expected ID=1460822101, got %q", ref.ID)
	}
}

func TestParsePageInput_ViewPageAction(t *testing.T) {
	input := "https://wiki.sits.no/pages/viewpage.action?pageId=99999"
	ref, err := parsePageInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ID != "99999" {
		t.Errorf("expected ID=99999, got %q", ref.ID)
	}
}

func TestParsePageInput_DisplayURL(t *testing.T) {
	input := "https://wiki.sits.no/display/DEV/Architecture+Overview"
	ref, err := parsePageInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ID != "" {
		t.Errorf("expected empty ID for display URL, got %q", ref.ID)
	}
	if ref.SpaceKey != "DEV" {
		t.Errorf("expected SpaceKey=DEV, got %q", ref.SpaceKey)
	}
	if ref.Title != "Architecture Overview" {
		t.Errorf("expected Title='Architecture Overview', got %q", ref.Title)
	}
}

func TestParsePageInput_DisplayURL_PercentEncoded(t *testing.T) {
	input := "https://wiki.sits.no/display/MUP/L%C3%B8sning+for+hemmeligheter"
	ref, err := parsePageInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.SpaceKey != "MUP" {
		t.Errorf("expected SpaceKey=MUP, got %q", ref.SpaceKey)
	}
	if ref.Title != "Løsning for hemmeligheter" {
		t.Errorf("expected Title='Løsning for hemmeligheter', got %q", ref.Title)
	}
}

func TestParsePageInput_PersonalSpaceURL(t *testing.T) {
	input := "https://wiki.sits.no/spaces/~k77319/pages/1481704261/Bruk+av+interne+data"
	ref, err := parsePageInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ID != "1481704261" {
		t.Errorf("expected ID=1481704261, got %q", ref.ID)
	}
}

func TestParsePageInput_UnknownURL_ReturnsError(t *testing.T) {
	input := "https://wiki.sits.no/some/unknown/path"
	_, err := parsePageInput(input)
	if err == nil {
		t.Fatal("expected error for unknown URL format")
	}
}

func TestParsePageInput_SpacesURLWithoutTitle(t *testing.T) {
	input := "https://wiki.sits.no/spaces/DEV/pages/12345"
	ref, err := parsePageInput(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ref.ID != "12345" {
		t.Errorf("expected ID=12345, got %q", ref.ID)
	}
}
