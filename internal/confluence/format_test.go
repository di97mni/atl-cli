package confluence

import (
	"strings"
	"testing"
)

func TestToMarkdown_BasicHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "paragraph",
			input:    "<p>Hello world</p>",
			expected: "Hello world",
		},
		{
			name:     "heading h1",
			input:    "<h1>Title</h1>",
			expected: "# Title",
		},
		{
			name:     "heading h2",
			input:    "<h2>Subtitle</h2>",
			expected: "## Subtitle",
		},
		{
			name:     "strong",
			input:    "<p><strong>bold text</strong></p>",
			expected: "**bold text**",
		},
		{
			name:     "emphasis",
			input:    "<p><em>italic text</em></p>",
			expected: "*italic text*",
		},
		{
			name:     "link",
			input:    `<p><a href="https://example.com">Example</a></p>`,
			expected: "[Example](https://example.com)",
		},
		{
			name:     "unordered list",
			input:    "<ul><li>Item 1</li><li>Item 2</li></ul>",
			expected: "- Item 1\n- Item 2",
		},
		{
			name:     "ordered list",
			input:    "<ol><li>First</li><li>Second</li></ol>",
			expected: "1. First\n2. Second",
		},
		{
			name:     "code inline",
			input:    "<p><code>inline code</code></p>",
			expected: "`inline code`",
		},
		{
			name:     "code block",
			input:    "<pre><code>code block</code></pre>",
			expected: "```\ncode block\n```",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToMarkdown(tt.input)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}
			result = strings.TrimSpace(result)
			if result != tt.expected {
				t.Errorf("ToMarkdown() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestToMarkdown_Tables(t *testing.T) {
	input := `<table>
		<thead>
			<tr><th>Column 1</th><th>Column 2</th></tr>
		</thead>
		<tbody>
			<tr><td>Value 1</td><td>Value 2</td></tr>
			<tr><td>Value 3</td><td>Value 4</td></tr>
		</tbody>
	</table>`

	result, err := ToMarkdown(input)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	// Check that the result contains table elements
	if !strings.Contains(result, "Column 1") {
		t.Error("expected result to contain 'Column 1'")
	}
	if !strings.Contains(result, "Column 2") {
		t.Error("expected result to contain 'Column 2'")
	}
	if !strings.Contains(result, "Value 1") {
		t.Error("expected result to contain 'Value 1'")
	}
	if !strings.Contains(result, "|") {
		t.Error("expected result to contain markdown table separators '|'")
	}
}

func TestToMarkdown_EntityDecoding(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ampersand",
			input:    "<p>Tom &amp; Jerry</p>",
			expected: "Tom & Jerry",
		},
		{
			// Less-than is kept escaped in markdown to avoid HTML interpretation
			name:     "less than",
			input:    "<p>1 &lt; 2</p>",
			expected: "1 &lt; 2",
		},
		{
			// Greater-than is kept escaped in markdown to avoid HTML interpretation
			name:     "greater than",
			input:    "<p>3 &gt; 2</p>",
			expected: "3 &gt; 2",
		},
		{
			// Smart quotes are converted to Unicode characters
			name:     "smart quotes",
			input:    "<p>&ldquo;quoted&rdquo;</p>",
			expected: "\u201cquoted\u201d", // Unicode left/right double quotation marks
		},
		{
			// Non-breaking space is preserved as Unicode
			name:     "nbsp",
			input:    "<p>hello&nbsp;world</p>",
			expected: "hello\u00a0world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToMarkdown(tt.input)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}
			result = strings.TrimSpace(result)
			if result != tt.expected {
				t.Errorf("ToMarkdown() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestToMarkdown_ConfluenceMacros(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "TOC macro",
			input:    `<ac:structured-macro ac:name="toc"><ac:parameter ac:name="maxLevel">3</ac:parameter></ac:structured-macro>`,
			expected: "[Table of Contents]",
		},
		{
			name:     "drawio macro with name",
			input:    `<ac:structured-macro ac:name="drawio"><ac:parameter ac:name="diagramName">Architecture Diagram</ac:parameter></ac:structured-macro>`,
			expected: "[Diagram: Architecture Diagram]",
		},
		{
			name:     "drawio macro without name",
			input:    `<ac:structured-macro ac:name="drawio"></ac:structured-macro>`,
			expected: "[Diagram]",
		},
		{
			name:     "code macro",
			input:    `<ac:structured-macro ac:name="code"><ac:parameter ac:name="language">python</ac:parameter><ac:plain-text-body><![CDATA[print("hello")]]></ac:plain-text-body></ac:structured-macro>`,
			expected: "```python\nprint(\"hello\")\n```",
		},
		{
			name:     "code macro without language",
			input:    `<ac:structured-macro ac:name="code"><ac:plain-text-body><![CDATA[some code]]></ac:plain-text-body></ac:structured-macro>`,
			expected: "```\nsome code\n```",
		},
		{
			name:     "info macro",
			input:    `<ac:structured-macro ac:name="info"><ac:rich-text-body><p>Important note</p></ac:rich-text-body></ac:structured-macro>`,
			expected: "> **Info:** Important note",
		},
		{
			name:     "warning macro",
			input:    `<ac:structured-macro ac:name="warning"><ac:rich-text-body><p>Be careful!</p></ac:rich-text-body></ac:structured-macro>`,
			expected: "> **Warning:** Be careful!",
		},
		{
			name:     "note macro",
			input:    `<ac:structured-macro ac:name="note"><ac:rich-text-body><p>A note here</p></ac:rich-text-body></ac:structured-macro>`,
			expected: "> **Note:** A note here",
		},
		{
			name:     "tip macro",
			input:    `<ac:structured-macro ac:name="tip"><ac:rich-text-body><p>A helpful tip</p></ac:rich-text-body></ac:structured-macro>`,
			expected: "> **Tip:** A helpful tip",
		},
		{
			name:     "unknown macro",
			input:    `<ac:structured-macro ac:name="custom-macro"><ac:parameter ac:name="foo">bar</ac:parameter></ac:structured-macro>`,
			expected: "[Confluence Macro: custom-macro]",
		},
		{
			name:     "expand macro",
			input:    `<ac:structured-macro ac:name="expand"><ac:parameter ac:name="title">Click to expand</ac:parameter><ac:rich-text-body><p>Hidden content</p></ac:rich-text-body></ac:structured-macro>`,
			expected: "<details>\n<summary>Click to expand</summary>\n\nHidden content\n</details>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToMarkdown(tt.input)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}
			result = strings.TrimSpace(result)
			if result != tt.expected {
				t.Errorf("ToMarkdown() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestToMarkdown_ConfluenceLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name:     "ac:link with ri:page",
			input:    `<ac:link><ri:page ri:content-title="Target Page" /><ac:plain-text-link-body><![CDATA[Link Text]]></ac:plain-text-link-body></ac:link>`,
			contains: "Link Text",
		},
		{
			name:     "ac:link without body",
			input:    `<ac:link><ri:page ri:content-title="Target Page" /></ac:link>`,
			contains: "Target Page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToMarkdown(tt.input)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}
			if !strings.Contains(result, tt.contains) {
				t.Errorf("ToMarkdown() = %q, expected to contain %q", result, tt.contains)
			}
		})
	}
}

func TestToMarkdown_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty body",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace only",
			input:    "   \n\t  ",
			expected: "",
		},
		{
			name:     "nested elements",
			input:    "<p><strong><em>bold and italic</em></strong></p>",
			expected: "***bold and italic***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ToMarkdown(tt.input)
			if err != nil {
				t.Fatalf("ToMarkdown() error = %v", err)
			}
			result = strings.TrimSpace(result)
			if result != tt.expected {
				t.Errorf("ToMarkdown() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestToMarkdown_ComplexDocument(t *testing.T) {
	input := `<ac:structured-macro ac:name="toc"></ac:structured-macro>
<p>Each Data Product is designed to be discoverable...</p>
<h1>Data Product Contents</h1>
<table><thead><tr><th>Column 1</th><th>Column 2</th></tr></thead><tbody><tr><td>Value 1</td><td>Value 2</td></tr></tbody></table>`

	result, err := ToMarkdown(input)
	if err != nil {
		t.Fatalf("ToMarkdown() error = %v", err)
	}

	// Verify key parts are present
	if !strings.Contains(result, "[Table of Contents]") {
		t.Error("expected TOC placeholder")
	}
	if !strings.Contains(result, "Each Data Product is designed to be discoverable") {
		t.Error("expected paragraph content")
	}
	if !strings.Contains(result, "# Data Product Contents") {
		t.Error("expected heading")
	}
	if !strings.Contains(result, "Column 1") {
		t.Error("expected table header")
	}
}
