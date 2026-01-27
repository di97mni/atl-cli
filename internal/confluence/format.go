package confluence

import (
	"regexp"
	"strings"

	"github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/table"
)

// ToMarkdown converts Confluence storage format (XHTML) to Markdown.
// It handles Confluence-specific macros and converts standard HTML elements.
func ToMarkdown(storageFormat string) (string, error) {
	if strings.TrimSpace(storageFormat) == "" {
		return "", nil
	}

	// Pre-process Confluence macros before HTML conversion
	processed := preprocessConfluenceMacros(storageFormat)

	// Create converter with plugins for full markdown support
	conv := converter.NewConverter(
		converter.WithPlugins(
			base.NewBasePlugin(),
			commonmark.NewCommonmarkPlugin(),
			table.NewTablePlugin(),
		),
	)

	// Convert HTML to Markdown
	result, err := conv.ConvertString(processed)
	if err != nil {
		return "", err
	}

	// Post-process to clean up
	result = postprocessMarkdown(result)

	return result, nil
}

// preprocessConfluenceMacros converts Confluence-specific macros to HTML
// that can be processed by the markdown converter.
func preprocessConfluenceMacros(html string) string {
	// Handle TOC macro - use special marker that won't be escaped
	html = regexp.MustCompile(`<ac:structured-macro[^>]*ac:name="toc"[^>]*>.*?</ac:structured-macro>`).
		ReplaceAllString(html, "<p>CFPLACEHOLDER:TOC:</p>")

	// Handle drawio macro - extract diagram name if present
	html = regexp.MustCompile(`<ac:structured-macro[^>]*ac:name="drawio"[^>]*>(?:.*?<ac:parameter[^>]*ac:name="diagramName"[^>]*>([^<]*)</ac:parameter>)?.*?</ac:structured-macro>`).
		ReplaceAllStringFunc(html, func(match string) string {
			nameMatch := regexp.MustCompile(`<ac:parameter[^>]*ac:name="diagramName"[^>]*>([^<]*)</ac:parameter>`).FindStringSubmatch(match)
			if len(nameMatch) > 1 && nameMatch[1] != "" {
				return "<p>CFPLACEHOLDER:DIAGRAM:" + nameMatch[1] + ":</p>"
			}
			return "<p>CFPLACEHOLDER:DIAGRAM::</p>"
		})

	// Handle code macro - convert to proper code block
	html = regexp.MustCompile(`<ac:structured-macro[^>]*ac:name="code"[^>]*>(.*?)</ac:structured-macro>`).
		ReplaceAllStringFunc(html, func(match string) string {
			// Extract language parameter
			lang := ""
			langMatch := regexp.MustCompile(`<ac:parameter[^>]*ac:name="language"[^>]*>([^<]*)</ac:parameter>`).FindStringSubmatch(match)
			if len(langMatch) > 1 {
				lang = langMatch[1]
			}

			// Extract code body
			code := ""
			codeMatch := regexp.MustCompile(`<ac:plain-text-body><!\[CDATA\[(.*?)\]\]></ac:plain-text-body>`).FindStringSubmatch(match)
			if len(codeMatch) > 1 {
				code = codeMatch[1]
			}

			return "<pre><code class=\"language-" + lang + "\">" + code + "</code></pre>"
		})

	// Handle info/warning/note/tip macros - convert to blockquotes
	for _, macroType := range []string{"info", "warning", "note", "tip"} {
		pattern := regexp.MustCompile(`<ac:structured-macro[^>]*ac:name="` + macroType + `"[^>]*>.*?<ac:rich-text-body>(.*?)</ac:rich-text-body>.*?</ac:structured-macro>`)
		label := strings.ToUpper(macroType[:1]) + macroType[1:]
		html = pattern.ReplaceAllStringFunc(html, func(match string) string {
			bodyMatch := regexp.MustCompile(`<ac:rich-text-body>(.*?)</ac:rich-text-body>`).FindStringSubmatch(match)
			if len(bodyMatch) > 1 {
				// Extract text from the body, removing HTML tags
				body := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(bodyMatch[1], "")
				body = strings.TrimSpace(body)
				return "<blockquote><strong>" + label + ":</strong> " + body + "</blockquote>"
			}
			return ""
		})
	}

	// Handle expand macro - convert to collapsible section format
	html = regexp.MustCompile(`<ac:structured-macro[^>]*ac:name="expand"[^>]*>(.*?)</ac:structured-macro>`).
		ReplaceAllStringFunc(html, func(match string) string {
			// Extract title parameter
			title := "Expand"
			titleMatch := regexp.MustCompile(`<ac:parameter[^>]*ac:name="title"[^>]*>([^<]*)</ac:parameter>`).FindStringSubmatch(match)
			if len(titleMatch) > 1 {
				title = titleMatch[1]
			}

			// Extract body content
			body := ""
			bodyMatch := regexp.MustCompile(`<ac:rich-text-body>(.*?)</ac:rich-text-body>`).FindStringSubmatch(match)
			if len(bodyMatch) > 1 {
				// Extract text from the body, removing HTML tags
				body = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(bodyMatch[1], "")
				body = strings.TrimSpace(body)
			}

			// Use a raw marker that will be converted in postprocessing
			return "CFPLACEHOLDER:EXPANDSTART:" + title + ":CFPLACEHOLDER:EXPANDBODY:" + body + ":CFPLACEHOLDER:EXPANDEND:"
		})

	// Handle ac:link elements
	html = regexp.MustCompile(`<ac:link>(.*?)</ac:link>`).
		ReplaceAllStringFunc(html, func(match string) string {
			// Try to extract link text from plain-text-link-body
			textMatch := regexp.MustCompile(`<ac:plain-text-link-body><!\[CDATA\[(.*?)\]\]></ac:plain-text-link-body>`).FindStringSubmatch(match)
			if len(textMatch) > 1 {
				return textMatch[1]
			}

			// Fall back to page title
			titleMatch := regexp.MustCompile(`ri:content-title="([^"]*)"`).FindStringSubmatch(match)
			if len(titleMatch) > 1 {
				return titleMatch[1]
			}

			return "CFPLACEHOLDER:LINK::"
		})

	// Handle unknown macros - replace with placeholder
	html = regexp.MustCompile(`<ac:structured-macro[^>]*ac:name="([^"]*)"[^>]*>.*?</ac:structured-macro>`).
		ReplaceAllStringFunc(html, func(match string) string {
			nameMatch := regexp.MustCompile(`ac:name="([^"]*)"`).FindStringSubmatch(match)
			if len(nameMatch) > 1 {
				return "<p>CFPLACEHOLDER:MACRO:" + nameMatch[1] + ":</p>"
			}
			return ""
		})

	return html
}

// postprocessMarkdown cleans up the converted markdown.
func postprocessMarkdown(markdown string) string {
	// Convert placeholders to final format
	markdown = strings.ReplaceAll(markdown, "CFPLACEHOLDER:TOC:", "[Table of Contents]")
	markdown = strings.ReplaceAll(markdown, "CFPLACEHOLDER:DIAGRAM::", "[Diagram]")
	markdown = strings.ReplaceAll(markdown, "CFPLACEHOLDER:LINK::", "[Link]")

	// Handle diagram placeholders with names
	markdown = regexp.MustCompile(`CFPLACEHOLDER:DIAGRAM:([^:]+):`).
		ReplaceAllString(markdown, "[Diagram: $1]")

	// Handle macro placeholders
	markdown = regexp.MustCompile(`CFPLACEHOLDER:MACRO:([^:]+):`).
		ReplaceAllString(markdown, "[Confluence Macro: $1]")

	// Handle expand/details placeholders
	markdown = regexp.MustCompile(`CFPLACEHOLDER:EXPANDSTART:([^:]*):CFPLACEHOLDER:EXPANDBODY:([^:]*):CFPLACEHOLDER:EXPANDEND:`).
		ReplaceAllString(markdown, "<details>\n<summary>$1</summary>\n\n$2\n</details>")

	// Normalize multiple newlines to max 2
	markdown = regexp.MustCompile(`\n{3,}`).ReplaceAllString(markdown, "\n\n")

	// Trim leading/trailing whitespace
	markdown = strings.TrimSpace(markdown)

	return markdown
}
