package formatter

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/bentsolheim/confluence-cli/internal/confluence"
)

const userMentionStorage = `<p>Hei <ac:link><ri:user ri:userkey="1f58644b4d9ba951014dae08fed80006" /></ac:link> og <ac:link><ri:user ri:username="k34046"/></ac:link>, og igjen <ac:link><ri:user ri:userkey="1f58644b4d9ba951014dae08fed80006" /></ac:link></p>`

func TestConvertBody_UserMention_Unresolved(t *testing.T) {
	got := ConvertBody(userMentionStorage)
	if !strings.Contains(got, "@user(1f58644b4d9ba951014dae08fed80006)") {
		t.Errorf("expected raw user key mention, got %q", got)
	}
	if !strings.Contains(got, "@user(k34046)") {
		t.Errorf("expected raw username mention, got %q", got)
	}
}

func TestConvertBodyWithUsers_ResolvesNames(t *testing.T) {
	names := map[string]string{
		"1f58644b4d9ba951014dae08fed80006": "Solheim, Bent Andre / k77319",
		"k34046":                           "Båtstrand, Anders / k34046",
	}
	got := ConvertBodyWithUsers(userMentionStorage, names)
	if strings.Count(got, "@user(Solheim, Bent Andre / k77319)") != 2 {
		t.Errorf("expected two resolved mentions, got %q", got)
	}
	if !strings.Contains(got, "@user(Båtstrand, Anders / k34046)") {
		t.Errorf("expected resolved username mention, got %q", got)
	}
	if strings.Contains(got, "1f58644b4d9ba951014dae08fed80006") {
		t.Errorf("expected user key to be replaced, got %q", got)
	}
}

func TestConvertBodyWithUsers_UnknownKeepsID(t *testing.T) {
	got := ConvertBodyWithUsers(userMentionStorage, map[string]string{"other": "Someone"})
	if !strings.Contains(got, "@user(1f58644b4d9ba951014dae08fed80006)") {
		t.Errorf("expected unknown user to keep its key, got %q", got)
	}
}

func TestExtractUserRefs(t *testing.T) {
	refs := ExtractUserRefs(userMentionStorage)
	if !reflect.DeepEqual(refs.Keys, []string{"1f58644b4d9ba951014dae08fed80006"}) {
		t.Errorf("unexpected keys: %v", refs.Keys)
	}
	if !reflect.DeepEqual(refs.Usernames, []string{"k34046"}) {
		t.Errorf("unexpected usernames: %v", refs.Usernames)
	}
}

func TestUserLabel(t *testing.T) {
	if got := UserLabel("Solheim, Bent Andre", "k77319"); got != "Solheim, Bent Andre / k77319" {
		t.Errorf("unexpected label: %q", got)
	}
	if got := UserLabel("", "k77319"); got != "k77319" {
		t.Errorf("unexpected label: %q", got)
	}
}

func TestMarkdownFormatter_UsesUserNames(t *testing.T) {
	f, err := New("markdown", "https://wiki.sits.no", Options{UserNames: map[string]string{
		"1f58644b4d9ba951014dae08fed80006": "Solheim, Bent Andre / k77319",
	}})
	if err != nil {
		t.Fatal(err)
	}
	page := &confluence.Page{ID: "1", Title: "T", Body: &confluence.Body{Storage: &confluence.BodyContent{Value: userMentionStorage}}}
	var buf bytes.Buffer
	if err := f.FormatPage(&buf, page); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "@user(Solheim, Bent Andre / k77319)") {
		t.Errorf("expected resolved name in markdown output, got %q", buf.String())
	}
}

func TestViewFormatter(t *testing.T) {
	f, err := New("view", "https://wiki.sits.no")
	if err != nil {
		t.Fatal(err)
	}
	page := &confluence.Page{ID: "1", Body: &confluence.Body{View: &confluence.BodyContent{Value: `<p><a href="/x" class="createlink">X</a></p>`}}}
	var buf bytes.Buffer
	if err := f.FormatPage(&buf, page); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "<p><a href=\"/x\" class=\"createlink\">X</a></p>\n" {
		t.Errorf("unexpected view output: %q", buf.String())
	}

	if err := f.FormatPage(&buf, &confluence.Page{ID: "2"}); err == nil {
		t.Error("expected error when view body is missing")
	}
}
