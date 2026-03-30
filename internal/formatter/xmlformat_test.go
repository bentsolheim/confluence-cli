package formatter

import (
	"strings"
	"testing"
)

func TestFormatStorageXML_SimpleTable(t *testing.T) {
	input := `<table class="wrapped"><tbody><tr><th scope="col">Date</th><th scope="col">Notes</th></tr><tr><td>2026-03-25</td><td>Some notes</td></tr></tbody></table>`
	got := FormatStorageXML(input)

	expected := `<table class="wrapped">
  <tbody>
    <tr>
      <th scope="col">Date</th>
      <th scope="col">Notes</th>
    </tr>
    <tr>
      <td>2026-03-25</td>
      <td>Some notes</td>
    </tr>
  </tbody>
</table>
`
	if got != expected {
		t.Errorf("table formatting mismatch.\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestFormatStorageXML_MacroWithCDATA(t *testing.T) {
	input := `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">go</ac:parameter><ac:plain-text-body><![CDATA[fmt.Println("hello")]]></ac:plain-text-body></ac:structured-macro>`
	got := FormatStorageXML(input)

	expected := `<ac:structured-macro ac:name="code">
  <ac:parameter ac:name="language">go</ac:parameter>
  <ac:plain-text-body><![CDATA[fmt.Println("hello")]]></ac:plain-text-body>
</ac:structured-macro>
`
	if got != expected {
		t.Errorf("macro formatting mismatch.\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestFormatStorageXML_InlineContent(t *testing.T) {
	input := `<p>Hello <strong>world</strong> and <em>friends</em>.</p>`
	got := FormatStorageXML(input)

	expected := `<p>
  Hello <strong>world</strong> and <em>friends</em>.
</p>
`
	if got != expected {
		t.Errorf("inline content mismatch.\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestFormatStorageXML_PreBlockPreserved(t *testing.T) {
	input := `<pre>  line 1
  line 2
  <b>bold</b>
</pre>`
	got := FormatStorageXML(input)

	// Pre content should be preserved exactly as-is
	expected := `<pre>  line 1
  line 2
  <b>bold</b>
</pre>
`
	if got != expected {
		t.Errorf("pre block mismatch.\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestFormatStorageXML_MixedBlockAndInline(t *testing.T) {
	input := `<p>Text with <a href="http://example.com">a link</a></p><p>Another paragraph</p>`
	got := FormatStorageXML(input)

	expected := `<p>
  Text with <a href="http://example.com">a link</a>
</p>
<p>
  Another paragraph
</p>
`
	if got != expected {
		t.Errorf("mixed content mismatch.\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestFormatStorageXML_NestedMacros(t *testing.T) {
	input := `<ac:structured-macro ac:name="expand"><ac:parameter ac:name="title">Details</ac:parameter><ac:rich-text-body><p>Hidden content</p></ac:rich-text-body></ac:structured-macro>`
	got := FormatStorageXML(input)

	expected := `<ac:structured-macro ac:name="expand">
  <ac:parameter ac:name="title">Details</ac:parameter>
  <ac:rich-text-body>
    <p>
      Hidden content
    </p>
  </ac:rich-text-body>
</ac:structured-macro>
`
	if got != expected {
		t.Errorf("nested macro mismatch.\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestFormatStorageXML_SelfClosingElements(t *testing.T) {
	input := `<p>Line 1<br />Line 2</p><hr />`
	got := FormatStorageXML(input)

	// br forces a newline at the same indent level, hr is block on its own line
	expected := `<p>
  Line 1<br />
  Line 2
</p>
<hr />
`
	if got != expected {
		t.Errorf("self-closing mismatch.\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestFormatStorageXML_ConfluenceLink(t *testing.T) {
	input := `<p>See <ac:link><ri:page ri:content-title="Other Page" /><ac:plain-text-link-body><![CDATA[Other Page]]></ac:plain-text-link-body></ac:link> for details.</p>`
	got := FormatStorageXML(input)

	// ac:link and ri:page are inline; p content is indented
	if !strings.Contains(got, "  See <ac:link>") {
		t.Errorf("expected indented link inside paragraph, got:\n%s", got)
	}
	if !strings.Contains(got, "<![CDATA[Other Page]]>") {
		t.Errorf("expected CDATA preserved in link, got:\n%s", got)
	}
}

func TestFormatStorageXML_Comment(t *testing.T) {
	input := `<p>Text</p><!-- comment --><p>More text</p>`
	got := FormatStorageXML(input)

	if !strings.Contains(got, "<!-- comment -->") {
		t.Errorf("expected comment preserved, got:\n%s", got)
	}
}

func TestFormatStorageXML_EmptyInput(t *testing.T) {
	if got := FormatStorageXML(""); got != "" {
		t.Errorf("expected empty output for empty input, got %q", got)
	}
	if got := FormatStorageXML("   "); got != "" {
		t.Errorf("expected empty output for whitespace input, got %q", got)
	}
}

func TestFormatStorageXML_Layout(t *testing.T) {
	input := `<ac:layout><ac:layout-section ac:type="two_right_sidebar"><ac:layout-cell><p>Main content</p></ac:layout-cell><ac:layout-cell><p>Sidebar</p></ac:layout-cell></ac:layout-section></ac:layout>`
	got := FormatStorageXML(input)

	expected := `<ac:layout>
  <ac:layout-section ac:type="two_right_sidebar">
    <ac:layout-cell>
      <p>
        Main content
      </p>
    </ac:layout-cell>
    <ac:layout-cell>
      <p>
        Sidebar
      </p>
    </ac:layout-cell>
  </ac:layout-section>
</ac:layout>
`
	if got != expected {
		t.Errorf("layout formatting mismatch.\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestFormatStorageXML_JiraMacro(t *testing.T) {
	input := `<ac:structured-macro ac:name="jira"><ac:parameter ac:name="server">Aurora JIRA</ac:parameter><ac:parameter ac:name="key">MUP-1192</ac:parameter></ac:structured-macro>`
	got := FormatStorageXML(input)

	if !strings.Contains(got, "MUP-1192") {
		t.Errorf("expected Jira key preserved, got:\n%s", got)
	}
	// Verify indentation
	if !strings.Contains(got, "  <ac:parameter") {
		t.Errorf("expected indented parameters, got:\n%s", got)
	}
}

func TestFormatStorageXML_CDataWithSpecialChars(t *testing.T) {
	input := `<ac:plain-text-body><![CDATA[if (a < b && c > d) { x = "hello"; }]]></ac:plain-text-body>`
	got := FormatStorageXML(input)

	// CDATA content must be preserved exactly, including < > & "
	if !strings.Contains(got, `<![CDATA[if (a < b && c > d) { x = "hello"; }]]>`) {
		t.Errorf("CDATA content was modified, got:\n%s", got)
	}
}

func TestFormatStorageXML_TaskList(t *testing.T) {
	input := `<ac:task-list><ac:task><ac:task-status>incomplete</ac:task-status><ac:task-body>Do something</ac:task-body></ac:task></ac:task-list>`
	got := FormatStorageXML(input)

	expected := `<ac:task-list>
  <ac:task>
    <ac:task-status>incomplete</ac:task-status>
    <ac:task-body>Do something</ac:task-body>
  </ac:task>
</ac:task-list>
`
	if got != expected {
		t.Errorf("task list mismatch.\nGot:\n%s\nExpected:\n%s", got, expected)
	}
}

func TestExtractTagName(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{`<p>`, "p"},
		{`</p>`, "p"},
		{`<br />`, "br"},
		{`<table class="wrapped">`, "table"},
		{`</ac:structured-macro>`, "ac:structured-macro"},
		{`<ac:parameter ac:name="key">`, "ac:parameter"},
		{`<ri:page ri:content-title="Title" />`, "ri:page"},
	}
	for _, tt := range tests {
		got := extractTagName(tt.input)
		if got != tt.want {
			t.Errorf("extractTagName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
