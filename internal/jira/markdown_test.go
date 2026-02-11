package jira

import (
	"encoding/json"
	"testing"
)

// --- Backward Compatibility Tests ---

func TestParseMarkdown_EmptyInput(t *testing.T) {
	result := ParseMarkdownToADFNodes("")
	if len(result) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(result))
	}
}

func TestParseMarkdown_WhitespaceOnly(t *testing.T) {
	result := ParseMarkdownToADFNodes("   \n\n  ")
	if len(result) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(result))
	}
}

func TestParseMarkdown_PlainText(t *testing.T) {
	result := ParseMarkdownToADFNodes("Hello world")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	assertParagraphText(t, result[0], "Hello world")
}

func TestParseMarkdown_MultipleParagraphs(t *testing.T) {
	result := ParseMarkdownToADFNodes("First paragraph\n\nSecond paragraph")
	if len(result) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result))
	}
	assertParagraphText(t, result[0], "First paragraph")
	assertParagraphText(t, result[1], "Second paragraph")
}

func TestParseMarkdown_SingleNewlinesCollapse(t *testing.T) {
	result := ParseMarkdownToADFNodes("Line one\nLine two")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	assertParagraphText(t, result[0], "Line one Line two")
}

// --- Heading Tests ---

func TestParseMarkdown_Headings(t *testing.T) {
	tests := []struct {
		input string
		level int
		text  string
	}{
		{"# H1", 1, "H1"},
		{"## H2", 2, "H2"},
		{"### H3", 3, "H3"},
		{"#### H4", 4, "H4"},
		{"##### H5", 5, "H5"},
		{"###### H6", 6, "H6"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result := ParseMarkdownToADFNodes(tc.input)
			if len(result) != 1 {
				t.Fatalf("expected 1 node, got %d", len(result))
			}
			node := result[0]
			if node.Type != "heading" {
				t.Fatalf("expected heading, got %q", node.Type)
			}
			level, ok := node.Attrs["level"]
			if !ok {
				t.Fatal("missing level attr")
			}
			if level != tc.level {
				t.Errorf("expected level %d, got %v", tc.level, level)
			}
			assertFirstTextContent(t, node, tc.text)
		})
	}
}

func TestParseMarkdown_HeadingWithInlineMarks(t *testing.T) {
	result := ParseMarkdownToADFNodes("## **Bold** heading")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	node := result[0]
	if node.Type != "heading" {
		t.Fatalf("expected heading, got %q", node.Type)
	}
	if len(node.Content) < 2 {
		t.Fatalf("expected at least 2 inline nodes, got %d", len(node.Content))
	}
	// First node: bold text.
	if node.Content[0].Text != "Bold" {
		t.Errorf("expected bold text 'Bold', got %q", node.Content[0].Text)
	}
	assertHasMark(t, node.Content[0], "strong")
}

func TestParseMarkdown_HeadingThenParagraph(t *testing.T) {
	result := ParseMarkdownToADFNodes("# Title\n\nBody text")
	if len(result) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result))
	}
	if result[0].Type != "heading" {
		t.Errorf("expected heading, got %q", result[0].Type)
	}
	assertParagraphText(t, result[1], "Body text")
}

func TestParseMarkdown_NotAHeading_NoSpace(t *testing.T) {
	result := ParseMarkdownToADFNodes("#nospace")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	if result[0].Type != "paragraph" {
		t.Errorf("expected paragraph, got %q", result[0].Type)
	}
}

func TestParseMarkdown_NotAHeading_TooManyHashes(t *testing.T) {
	result := ParseMarkdownToADFNodes("####### Seven")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	if result[0].Type != "paragraph" {
		t.Errorf("expected paragraph, got %q", result[0].Type)
	}
}

// --- Inline Formatting Tests ---

func TestParseInline_Bold(t *testing.T) {
	nodes := parseInline("**bold**")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Text != "bold" {
		t.Errorf("expected 'bold', got %q", nodes[0].Text)
	}
	assertHasMark(t, nodes[0], "strong")
}

func TestParseInline_BoldInContext(t *testing.T) {
	nodes := parseInline("before **bold** after")
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	if nodes[0].Text != "before " {
		t.Errorf("expected 'before ', got %q", nodes[0].Text)
	}
	if nodes[1].Text != "bold" {
		t.Errorf("expected 'bold', got %q", nodes[1].Text)
	}
	assertHasMark(t, nodes[1], "strong")
	if nodes[2].Text != " after" {
		t.Errorf("expected ' after', got %q", nodes[2].Text)
	}
}

func TestParseInline_ItalicStar(t *testing.T) {
	nodes := parseInline("*italic*")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Text != "italic" {
		t.Errorf("expected 'italic', got %q", nodes[0].Text)
	}
	assertHasMark(t, nodes[0], "em")
}

func TestParseInline_ItalicUnderscore(t *testing.T) {
	nodes := parseInline("_italic_")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Text != "italic" {
		t.Errorf("expected 'italic', got %q", nodes[0].Text)
	}
	assertHasMark(t, nodes[0], "em")
}

func TestParseInline_BoldItalic(t *testing.T) {
	nodes := parseInline("***both***")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Text != "both" {
		t.Errorf("expected 'both', got %q", nodes[0].Text)
	}
	assertHasMark(t, nodes[0], "strong")
	assertHasMark(t, nodes[0], "em")
}

func TestParseInline_CodeSpan(t *testing.T) {
	nodes := parseInline("`code`")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Text != "code" {
		t.Errorf("expected 'code', got %q", nodes[0].Text)
	}
	assertHasMark(t, nodes[0], "code")
}

func TestParseInline_CodePreservesInner(t *testing.T) {
	nodes := parseInline("`**not bold**`")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Text != "**not bold**" {
		t.Errorf("expected '**not bold**', got %q", nodes[0].Text)
	}
	assertHasMark(t, nodes[0], "code")
	assertNoMark(t, nodes[0], "strong")
}

func TestParseInline_Link(t *testing.T) {
	nodes := parseInline("[text](https://example.com)")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Text != "text" {
		t.Errorf("expected 'text', got %q", nodes[0].Text)
	}
	assertHasMark(t, nodes[0], "link")
	// Verify href.
	for _, mark := range nodes[0].Marks {
		if mark.Type == "link" {
			href, ok := mark.Attrs["href"]
			if !ok {
				t.Error("link mark missing href attr")
			} else if href != "https://example.com" {
				t.Errorf("expected href 'https://example.com', got %v", href)
			}
		}
	}
}

func TestParseInline_LinkInContext(t *testing.T) {
	nodes := parseInline("See [docs](url) here")
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}
	if nodes[0].Text != "See " {
		t.Errorf("expected 'See ', got %q", nodes[0].Text)
	}
	if nodes[1].Text != "docs" {
		t.Errorf("expected 'docs', got %q", nodes[1].Text)
	}
	assertHasMark(t, nodes[1], "link")
	if nodes[2].Text != " here" {
		t.Errorf("expected ' here', got %q", nodes[2].Text)
	}
}

func TestParseInline_MultipleMixed(t *testing.T) {
	nodes := parseInline("**bold** and *italic*")
	// Expect: bold node, " and " text, italic node.
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d: %+v", len(nodes), nodes)
	}
	assertHasMark(t, nodes[0], "strong")
	assertHasMark(t, nodes[2], "em")
}

func TestParseInline_NoMarkers(t *testing.T) {
	nodes := parseInline("plain text")
	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}
	if nodes[0].Text != "plain text" {
		t.Errorf("expected 'plain text', got %q", nodes[0].Text)
	}
	if len(nodes[0].Marks) != 0 {
		t.Errorf("expected no marks, got %d", len(nodes[0].Marks))
	}
}

// --- List Tests ---

func TestParseMarkdown_BulletListDash(t *testing.T) {
	result := ParseMarkdownToADFNodes("- one\n- two")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	assertListType(t, result[0], "bulletList", 2)
}

func TestParseMarkdown_BulletListStar(t *testing.T) {
	result := ParseMarkdownToADFNodes("* one\n* two")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	assertListType(t, result[0], "bulletList", 2)
}

func TestParseMarkdown_OrderedList(t *testing.T) {
	result := ParseMarkdownToADFNodes("1. first\n2. second\n3. third")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	assertListType(t, result[0], "orderedList", 3)
}

func TestParseMarkdown_ListWithInlineMarks(t *testing.T) {
	result := ParseMarkdownToADFNodes("- **bold** item")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	if result[0].Type != "bulletList" {
		t.Fatalf("expected bulletList, got %q", result[0].Type)
	}
	// Navigate: bulletList -> listItem -> paragraph -> first inline node.
	listItem := result[0].Content[0]
	para := listItem.Content[0]
	if len(para.Content) < 2 {
		t.Fatalf("expected at least 2 inline nodes, got %d", len(para.Content))
	}
	assertHasMark(t, para.Content[0], "strong")
}

func TestParseMarkdown_BlankLineBreaksList(t *testing.T) {
	result := ParseMarkdownToADFNodes("- a\n\n- b")
	if len(result) != 2 {
		t.Fatalf("expected 2 nodes (2 separate lists), got %d", len(result))
	}
	assertListType(t, result[0], "bulletList", 1)
	assertListType(t, result[1], "bulletList", 1)
}

func TestParseMarkdown_ListTypeSwitchFlushes(t *testing.T) {
	result := ParseMarkdownToADFNodes("- bullet\n1. ordered")
	if len(result) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result))
	}
	assertListType(t, result[0], "bulletList", 1)
	assertListType(t, result[1], "orderedList", 1)
}

func TestParseMarkdown_ParagraphThenList(t *testing.T) {
	result := ParseMarkdownToADFNodes("Some text\n\n- item one\n- item two")
	if len(result) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(result))
	}
	assertParagraphText(t, result[0], "Some text")
	assertListType(t, result[1], "bulletList", 2)
}

// --- Code Block Tests ---

func TestParseMarkdown_CodeBlock(t *testing.T) {
	result := ParseMarkdownToADFNodes("```\ncode\n```")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	if result[0].Type != "codeBlock" {
		t.Fatalf("expected codeBlock, got %q", result[0].Type)
	}
	assertCodeBlockText(t, result[0], "code")
}

func TestParseMarkdown_CodeBlockWithLanguage(t *testing.T) {
	result := ParseMarkdownToADFNodes("```go\nfunc main() {}\n```")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	node := result[0]
	if node.Type != "codeBlock" {
		t.Fatalf("expected codeBlock, got %q", node.Type)
	}
	lang, ok := node.Attrs["language"]
	if !ok {
		t.Fatal("missing language attr")
	}
	if lang != "go" {
		t.Errorf("expected language 'go', got %v", lang)
	}
	assertCodeBlockText(t, node, "func main() {}")
}

func TestParseMarkdown_CodeBlockMultiLine(t *testing.T) {
	input := "```\nline1\nline2\nline3\n```"
	result := ParseMarkdownToADFNodes(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	assertCodeBlockText(t, result[0], "line1\nline2\nline3")
}

func TestParseMarkdown_CodeBlockPreservesMarkdown(t *testing.T) {
	input := "```\n**not bold**\n# not heading\n```"
	result := ParseMarkdownToADFNodes(input)
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	assertCodeBlockText(t, result[0], "**not bold**\n# not heading")
}

func TestParseMarkdown_UnclosedCodeBlock(t *testing.T) {
	result := ParseMarkdownToADFNodes("```\ncode")
	if len(result) != 1 {
		t.Fatalf("expected 1 node, got %d", len(result))
	}
	if result[0].Type != "codeBlock" {
		t.Fatalf("expected codeBlock, got %q", result[0].Type)
	}
	assertCodeBlockText(t, result[0], "code")
}

// --- Integration Tests ---

func TestParseMarkdown_MixedDocument(t *testing.T) {
	input := "# Title\n\nSome **bold** text.\n\n- item one\n- item two\n\n```go\nfmt.Println(\"hi\")\n```"
	result := ParseMarkdownToADFNodes(input)

	if len(result) != 4 {
		t.Fatalf("expected 4 nodes (heading, paragraph, list, code), got %d", len(result))
	}
	if result[0].Type != "heading" {
		t.Errorf("node 0: expected heading, got %q", result[0].Type)
	}
	if result[1].Type != "paragraph" {
		t.Errorf("node 1: expected paragraph, got %q", result[1].Type)
	}
	if result[2].Type != "bulletList" {
		t.Errorf("node 2: expected bulletList, got %q", result[2].Type)
	}
	if result[3].Type != "codeBlock" {
		t.Errorf("node 3: expected codeBlock, got %q", result[3].Type)
	}
}

func TestTextToADF_MarkdownHeading_JSON(t *testing.T) {
	doc := TextToADF("## Hello")
	if doc == nil {
		t.Fatal("expected non-nil doc")
	}
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	if parsed["type"] != "doc" {
		t.Errorf("expected type 'doc', got %v", parsed["type"])
	}

	content, ok := parsed["content"].([]interface{})
	if !ok || len(content) != 1 {
		t.Fatalf("expected 1 content node, got %v", parsed["content"])
	}

	heading, ok := content[0].(map[string]interface{})
	if !ok {
		t.Fatal("expected heading to be a map")
	}
	if heading["type"] != "heading" {
		t.Errorf("expected heading type, got %v", heading["type"])
	}

	attrs, ok := heading["attrs"].(map[string]interface{})
	if !ok {
		t.Fatal("expected attrs to be a map")
	}
	if attrs["level"] != float64(2) {
		t.Errorf("expected level 2, got %v", attrs["level"])
	}
}

// Verify that TextToADF backward compatibility with plain text is maintained.
func TestTextToADF_PlainTextBackwardCompat(t *testing.T) {
	doc := TextToADF("First paragraph\n\nSecond paragraph")
	if doc == nil {
		t.Fatal("expected non-nil doc")
	}
	if doc.Type != "doc" {
		t.Errorf("expected type 'doc', got %q", doc.Type)
	}
	if doc.Version != 1 {
		t.Errorf("expected version 1, got %d", doc.Version)
	}
	if len(doc.Content) != 2 {
		t.Fatalf("expected 2 paragraphs, got %d", len(doc.Content))
	}
	assertParagraphText(t, doc.Content[0], "First paragraph")
	assertParagraphText(t, doc.Content[1], "Second paragraph")
}

// --- Test Helpers ---

func assertParagraphText(t *testing.T, node ADFNode, expected string) {
	t.Helper()
	if node.Type != "paragraph" {
		t.Errorf("expected paragraph, got %q", node.Type)
		return
	}
	if len(node.Content) != 1 {
		t.Errorf("expected 1 text node in paragraph, got %d", len(node.Content))
		return
	}
	if node.Content[0].Text != expected {
		t.Errorf("expected text %q, got %q", expected, node.Content[0].Text)
	}
}

func assertFirstTextContent(t *testing.T, node ADFNode, expected string) {
	t.Helper()
	if len(node.Content) == 0 {
		t.Error("expected content, got none")
		return
	}
	if node.Content[0].Text != expected {
		t.Errorf("expected first text %q, got %q", expected, node.Content[0].Text)
	}
}

func assertHasMark(t *testing.T, node ADFNode, markType string) {
	t.Helper()
	for _, m := range node.Marks {
		if m.Type == markType {
			return
		}
	}
	t.Errorf("expected mark %q on node with text %q, not found", markType, node.Text)
}

func assertNoMark(t *testing.T, node ADFNode, markType string) {
	t.Helper()
	for _, m := range node.Marks {
		if m.Type == markType {
			t.Errorf("expected no mark %q on node with text %q, but found it", markType, node.Text)
			return
		}
	}
}

func assertListType(t *testing.T, node ADFNode, expectedType string, expectedItems int) {
	t.Helper()
	if node.Type != expectedType {
		t.Errorf("expected %s, got %q", expectedType, node.Type)
		return
	}
	if len(node.Content) != expectedItems {
		t.Errorf("expected %d list items, got %d", expectedItems, len(node.Content))
		return
	}
	for i, item := range node.Content {
		if item.Type != "listItem" {
			t.Errorf("item %d: expected listItem, got %q", i, item.Type)
		}
	}
}

func assertCodeBlockText(t *testing.T, node ADFNode, expected string) {
	t.Helper()
	if node.Type != "codeBlock" {
		t.Errorf("expected codeBlock, got %q", node.Type)
		return
	}
	if len(node.Content) == 0 {
		if expected != "" {
			t.Errorf("expected code text %q, got empty codeBlock", expected)
		}
		return
	}
	if node.Content[0].Text != expected {
		t.Errorf("expected code text %q, got %q", expected, node.Content[0].Text)
	}
}
