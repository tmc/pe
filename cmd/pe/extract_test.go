package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestExtractCmd_FlagParsing(t *testing.T) {
	cmd := extractCmd()

	// Test flag existence
	flags := []string{"tag", "tags", "all", "format", "xpath", "nested", "validate", "attr", "transform", "output", "start", "end", "stream"}
	for _, name := range flags {
		if cmd.Flags().Lookup(name) == nil {
			t.Errorf("Expected flag %q to exist", name)
		}
	}
}

func TestExtractCmd_CommandStructure(t *testing.T) {
	cmd := extractCmd()

	if cmd.Use != "extract" {
		t.Errorf("Unexpected Use: %s", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected Short description to be set")
	}
}

func TestExtractCmdOutputDeniedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("input.txt", []byte("<answer>42</answer>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("pe.mod", []byte(extractPolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := extractCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"input.txt", "--tag", "answer", "--output", "answer.txt"})

	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "tool write is denied") {
		t.Fatalf("extract error = %v, want write policy denial", err)
	}
	if _, err := os.Stat("answer.txt"); !os.IsNotExist(err) {
		t.Fatalf("extract wrote output despite write policy: %v", err)
	}
}

func TestExtractCmdStdoutAllowedByWritePolicy(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(oldWd)

	if err := os.WriteFile("input.txt", []byte("<answer>42</answer>"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("pe.mod", []byte(extractPolicyTestModule()), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := extractCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"input.txt", "--tag", "answer"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("extract stdout: %v", err)
	}
	if got := buf.String(); got != "42\n" {
		t.Fatalf("extract stdout = %q, want 42 newline", got)
	}
	if _, err := os.Stat("answer.txt"); !os.IsNotExist(err) {
		t.Fatalf("extract created unexpected output: %v", err)
	}
}

func extractPolicyTestModule() string {
	return `module example.com/prompts

pe 1

capability {
    tools deny write
}
`
}

func TestExtractTags_SimpleTag(t *testing.T) {
	content := `<answer>This is the answer</answer>`
	results, err := extractTags(content, "answer", false, "", false, "")
	if err != nil {
		t.Errorf("extractTags() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("extractTags() got %d results, want 1", len(results))
	}
	if results[0].Content != "This is the answer" {
		t.Errorf("extractTags() content = %q, want %q", results[0].Content, "This is the answer")
	}
}

func TestExtractTags_AllMatches(t *testing.T) {
	content := `<item>First</item><item>Second</item><item>Third</item>`
	results, err := extractTags(content, "item", true, "", false, "")
	if err != nil {
		t.Errorf("extractTags() error = %v", err)
	}
	if len(results) != 3 {
		t.Errorf("extractTags() got %d results, want 3", len(results))
	}
}

func TestExtractTags_MultilineContent(t *testing.T) {
	content := `<code>
function hello() {
    return "world";
}
</code>`
	results, err := extractTags(content, "code", false, "", false, "")
	if err != nil {
		t.Errorf("extractTags() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("extractTags() got %d results, want 1", len(results))
	}
	if !strings.Contains(results[0].Content, "function hello()") {
		t.Errorf("extractTags() content missing function definition")
	}
}

func TestExtractTags_TagNotFound(t *testing.T) {
	content := `<answer>content</answer>`
	_, err := extractTags(content, "nonexistent", false, "", false, "")
	if err == nil {
		t.Error("extractTags() should error for nonexistent tag")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("extractTags() error = %q, should contain 'not found'", err.Error())
	}
}

func TestExtractTags_EmptyTag(t *testing.T) {
	content := `<answer>content</answer>`
	_, err := extractTags(content, "", false, "", false, "")
	if err == nil {
		t.Error("extractTags() should error for empty tag")
	}
}

func TestExtractTags_WithAttribute(t *testing.T) {
	content := `<answer type="final">This is it</answer><answer type="draft">Maybe this</answer>`
	results, err := extractTags(content, "answer", false, "", false, "type=final")
	if err != nil {
		t.Errorf("extractTags() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("extractTags() got %d results, want 1", len(results))
	}
	if results[0].Content != "This is it" {
		t.Errorf("extractTags() content = %q, want %q", results[0].Content, "This is it")
	}
}

func TestExtractCustomDelimiters_Simple(t *testing.T) {
	content := `<<<START>>>Content here<<<END>>>`
	results, err := extractCustomDelimiters(content, "<<<START>>>", "<<<END>>>", false)
	if err != nil {
		t.Errorf("extractCustomDelimiters() error = %v", err)
	}
	if len(results) != 1 {
		t.Errorf("extractCustomDelimiters() got %d results, want 1", len(results))
	}
	if results[0].Content != "Content here" {
		t.Errorf("extractCustomDelimiters() content = %q, want %q", results[0].Content, "Content here")
	}
}

func TestExtractCustomDelimiters_NotFound(t *testing.T) {
	content := `No delimiters here`
	_, err := extractCustomDelimiters(content, "<<<START>>>", "<<<END>>>", false)
	if err == nil {
		t.Error("extractCustomDelimiters() should error when delimiters not found")
	}
}

func TestExtractMultipleTags(t *testing.T) {
	content := `<thinking>some thought</thinking><answer>the answer</answer>`
	results, err := extractMultipleTags(content, "thinking,answer", true, false, "")
	if err != nil {
		t.Errorf("extractMultipleTags() error = %v", err)
	}
	if len(results) < 2 {
		t.Errorf("extractMultipleTags() got %d results, want at least 2", len(results))
	}
}

func TestApplyTransform_NoOp(t *testing.T) {
	results := []ExtractedContent{
		{Content: "hello"},
	}
	transformed, err := applyTransform(results, "none")
	if err != nil {
		t.Errorf("applyTransform() error = %v", err)
	}
	if transformed[0].Content != "hello" {
		t.Errorf("applyTransform() content = %q, want %q", transformed[0].Content, "hello")
	}
}

func TestApplyTransform_Fmt(t *testing.T) {
	results := []ExtractedContent{
		{Content: "func main() {\n  fmt.Println(\"hello\")\n}"},
	}
	transformed, err := applyTransform(results, "fmt")
	if err != nil {
		t.Errorf("applyTransform() error = %v", err)
	}
	if transformed[0].Content == "" {
		t.Error("applyTransform(fmt) returned empty content")
	}
}

func TestExtractedContentStruct(t *testing.T) {
	ec := ExtractedContent{
		Tag:     "answer",
		Content: "test content",
		Attrs:   map[string]string{"type": "final"},
		Path:    "/root/answer",
	}

	if ec.Tag != "answer" {
		t.Error("ExtractedContent.Tag mismatch")
	}
	if ec.Content != "test content" {
		t.Error("ExtractedContent.Content mismatch")
	}
}
