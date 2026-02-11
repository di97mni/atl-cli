package jira

import (
	"regexp"
	"strings"
)

// Block-level patterns.
var (
	headingRe     = regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	bulletItemRe  = regexp.MustCompile(`^[-*]\s+(.+)$`)
	orderedItemRe = regexp.MustCompile(`^\d+\.\s+(.+)$`)
	codeFenceRe   = regexp.MustCompile("^```(\\w*)\\s*$")
)

// Inline patterns (order matters: longest delimiter first).
var (
	boldItalicRe  = regexp.MustCompile(`\*\*\*(.*?)\*\*\*`)
	boldRe        = regexp.MustCompile(`\*\*(.*?)\*\*`)
	italicStarRe  = regexp.MustCompile(`(?:^|[^*])\*([^*]+?)\*(?:[^*]|$)`)
	italicUnderRe = regexp.MustCompile(`(?:^|[^_])_([^_]+?)_(?:[^_]|$)`)
	codeSpanRe    = regexp.MustCompile("`([^`]+)`")
	linkRe        = regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
)

// inlineMatch tracks a single inline pattern match within a text string.
type inlineMatch struct {
	start int     // start offset of the full match in the source text
	end   int     // end offset (exclusive) of the full match
	node  ADFNode // the ADF node to emit for this match
}

// ParseMarkdownToADFNodes parses markdown text into a slice of block-level ADF nodes.
func ParseMarkdownToADFNodes(text string) []ADFNode {
	var result []ADFNode

	var paragraphLines []string
	var listType string
	var listItems [][]ADFNode
	var inCodeBlock bool
	var codeBlockLang string
	var codeBlockLines []string

	flushParagraph := func() {
		if len(paragraphLines) == 0 {
			return
		}
		joined := strings.Join(paragraphLines, " ")
		joined = strings.TrimSpace(joined)
		if joined != "" {
			result = append(result, makeParagraph(parseInline(joined)))
		}
		paragraphLines = nil
	}

	flushList := func() {
		if listType == "" || len(listItems) == 0 {
			listType = ""
			listItems = nil
			return
		}
		result = append(result, makeList(listType, listItems))
		listType = ""
		listItems = nil
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Inside a code block: accumulate lines until closing fence.
		if inCodeBlock {
			if codeFenceRe.MatchString(trimmed) && trimmed == "```" {
				result = append(result, makeCodeBlock(codeBlockLang, codeBlockLines))
				inCodeBlock = false
				codeBlockLang = ""
				codeBlockLines = nil
			} else {
				codeBlockLines = append(codeBlockLines, line)
			}
			continue
		}

		// Opening code fence.
		if m := codeFenceRe.FindStringSubmatch(trimmed); m != nil {
			flushParagraph()
			flushList()
			inCodeBlock = true
			codeBlockLang = m[1]
			codeBlockLines = nil
			continue
		}

		// Empty line: flush current blocks.
		if trimmed == "" {
			flushParagraph()
			flushList()
			continue
		}

		// Heading.
		if m := headingRe.FindStringSubmatch(trimmed); m != nil {
			flushParagraph()
			flushList()
			level := len(m[1])
			content := strings.TrimSpace(m[2])
			result = append(result, makeHeading(level, parseInline(content)))
			continue
		}

		// Bullet list item.
		if m := bulletItemRe.FindStringSubmatch(trimmed); m != nil {
			flushParagraph()
			if listType != "" && listType != "bulletList" {
				flushList()
			}
			listType = "bulletList"
			listItems = append(listItems, parseInline(m[1]))
			continue
		}

		// Ordered list item.
		if m := orderedItemRe.FindStringSubmatch(trimmed); m != nil {
			flushParagraph()
			if listType != "" && listType != "orderedList" {
				flushList()
			}
			listType = "orderedList"
			listItems = append(listItems, parseInline(m[1]))
			continue
		}

		// Default: paragraph continuation line.
		flushList()
		paragraphLines = append(paragraphLines, trimmed)
	}

	// Flush remaining state.
	flushParagraph()
	flushList()
	if inCodeBlock {
		result = append(result, makeCodeBlock(codeBlockLang, codeBlockLines))
	}

	return result
}

// parseInline parses inline markdown formatting within a text string.
// It returns a slice of ADF text nodes, with marks applied for bold, italic, code, and links.
func parseInline(text string) []ADFNode {
	var nodes []ADFNode
	for text != "" {
		m := findEarliestInlineMatch(text)
		if m == nil {
			nodes = append(nodes, makeTextNode(text))
			break
		}
		if m.start > 0 {
			nodes = append(nodes, makeTextNode(text[:m.start]))
		}
		nodes = append(nodes, m.node)
		text = text[m.end:]
	}
	return nodes
}

// findEarliestInlineMatch finds the first inline markdown pattern in text.
// Returns nil if no patterns match.
func findEarliestInlineMatch(text string) *inlineMatch {
	var best *inlineMatch

	tryPattern := func(re *regexp.Regexp, build func([]int) ADFNode, matchStart, matchEnd, captureStart, captureEnd int) {
		loc := re.FindStringSubmatchIndex(text)
		if loc == nil {
			return
		}
		// Check for empty captured content.
		cs, ce := loc[captureStart], loc[captureEnd]
		if ce <= cs {
			return
		}
		start, end := loc[matchStart], loc[matchEnd]
		if best == nil || start < best.start || (start == best.start && end > best.end) {
			best = &inlineMatch{start: start, end: end, node: build(loc)}
		}
	}

	// Standard patterns: full match is group 0 (indices 0,1), capture is group 1 (indices 2,3).
	std := func(re *regexp.Regexp, build func(string) ADFNode) {
		tryPattern(re, func(loc []int) ADFNode {
			return build(text[loc[2]:loc[3]])
		}, 0, 1, 2, 3)
	}

	// Patterns with context chars that need offset adjustment.
	// italicStarRe and italicUnderRe use lookbehind/lookahead context chars.
	// The actual match to replace is from 1 char after start to 1 char before end
	// (when context chars are present).
	italicTry := func(re *regexp.Regexp) {
		loc := re.FindStringSubmatchIndex(text)
		if loc == nil {
			return
		}
		// Group 1 is the captured italic text.
		cs, ce := loc[2], loc[3]
		if ce <= cs {
			return
		}
		captured := text[cs:ce]

		// Find the actual * or _ delimiters around the captured group.
		// The delimiter is 1 char before capture start and 1 char after capture end.
		delimStart := cs - 1 // position of opening * or _
		delimEnd := ce + 1   // position after closing * or _

		if best == nil || delimStart < best.start || (delimStart == best.start && delimEnd > best.end) {
			best = &inlineMatch{
				start: delimStart,
				end:   delimEnd,
				node:  makeMarkedText(captured, []ADFMark{{Type: "em"}}),
			}
		}
	}

	// Order: longest delimiter first to resolve ambiguity.
	std(boldItalicRe, func(s string) ADFNode {
		return makeMarkedText(s, []ADFMark{{Type: "strong"}, {Type: "em"}})
	})
	std(boldRe, func(s string) ADFNode {
		return makeMarkedText(s, []ADFMark{{Type: "strong"}})
	})
	italicTry(italicStarRe)
	italicTry(italicUnderRe)
	std(codeSpanRe, func(s string) ADFNode {
		return makeMarkedText(s, []ADFMark{{Type: "code"}})
	})

	// Link: group 1 = text, group 2 = href.
	tryPattern(linkRe, func(loc []int) ADFNode {
		linkText := text[loc[2]:loc[3]]
		href := text[loc[4]:loc[5]]
		return makeMarkedText(linkText, []ADFMark{
			{Type: "link", Attrs: map[string]interface{}{"href": href}},
		})
	}, 0, 1, 2, 3)

	return best
}

// Helper constructors for ADF nodes.

func makeTextNode(text string) ADFNode {
	return ADFNode{Type: "text", Text: text}
}

func makeMarkedText(text string, marks []ADFMark) ADFNode {
	return ADFNode{Type: "text", Text: text, Marks: marks}
}

func makeParagraph(content []ADFNode) ADFNode {
	return ADFNode{Type: "paragraph", Content: content}
}

func makeHeading(level int, content []ADFNode) ADFNode {
	return ADFNode{
		Type:    "heading",
		Attrs:   map[string]interface{}{"level": level},
		Content: content,
	}
}

func makeList(listType string, items [][]ADFNode) ADFNode {
	listItems := make([]ADFNode, len(items))
	for i, itemContent := range items {
		listItems[i] = ADFNode{
			Type: "listItem",
			Content: []ADFNode{
				makeParagraph(itemContent),
			},
		}
	}
	return ADFNode{Type: listType, Content: listItems}
}

func makeCodeBlock(language string, lines []string) ADFNode {
	node := ADFNode{Type: "codeBlock"}
	if language != "" {
		node.Attrs = map[string]interface{}{"language": language}
	}
	text := strings.Join(lines, "\n")
	if text != "" {
		node.Content = []ADFNode{makeTextNode(text)}
	}
	return node
}
