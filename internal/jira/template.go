package jira

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

// TemplateFrontmatter contains the YAML frontmatter metadata.
type TemplateFrontmatter struct {
	Version   int      `yaml:"version"`
	IssueType string   `yaml:"issueType"`
	Project   string   `yaml:"project"`
	Summary   string   `yaml:"summary"`
	Labels    []string `yaml:"labels"`
}

// Template represents a parsed issue template.
type Template struct {
	Frontmatter TemplateFrontmatter
	Body        string
}

// ParsedTemplate contains the fully processed template with variables applied.
type ParsedTemplate struct {
	Project     string
	IssueType   string
	Summary     string
	Labels      []string
	Description string
}

// LoadTemplate loads and parses a template file.
func LoadTemplate(path string) (*Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	return ParseTemplate(string(content))
}

// ParseTemplate parses template content into frontmatter and body.
func ParseTemplate(content string) (*Template, error) {
	// Check for YAML frontmatter delimiters
	if !strings.HasPrefix(content, "---\n") {
		return nil, fmt.Errorf("template must start with YAML frontmatter (---)")
	}

	// Find the closing delimiter
	rest := content[4:] // Skip opening "---\n"
	endIdx := strings.Index(rest, "\n---")
	if endIdx == -1 {
		return nil, fmt.Errorf("template frontmatter not properly closed (missing ---)")
	}

	frontmatterYAML := rest[:endIdx]
	body := strings.TrimPrefix(rest[endIdx+4:], "\n") // Skip "\n---" and optional newline

	// Parse YAML frontmatter
	var fm TemplateFrontmatter
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &fm); err != nil {
		return nil, fmt.Errorf("failed to parse template frontmatter: %w", err)
	}

	// Validate required fields
	if fm.Version != 1 {
		return nil, fmt.Errorf("unsupported template version: %d (expected 1)", fm.Version)
	}

	return &Template{
		Frontmatter: fm,
		Body:        body,
	}, nil
}

// Apply applies variables to the template and returns a ParsedTemplate.
func (t *Template) Apply(vars map[string]string) (*ParsedTemplate, error) {
	// Apply variables to summary
	summary, err := applyTemplateVars(t.Frontmatter.Summary, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to apply variables to summary: %w", err)
	}

	// Apply variables to body (description)
	description, err := applyTemplateVars(t.Body, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to apply variables to description: %w", err)
	}

	// Apply variables to labels
	labels := make([]string, 0, len(t.Frontmatter.Labels))
	for _, label := range t.Frontmatter.Labels {
		processedLabel, err := applyTemplateVars(label, vars)
		if err != nil {
			return nil, fmt.Errorf("failed to apply variables to label: %w", err)
		}
		if processedLabel != "" {
			labels = append(labels, processedLabel)
		}
	}

	return &ParsedTemplate{
		Project:     t.Frontmatter.Project,
		IssueType:   t.Frontmatter.IssueType,
		Summary:     summary,
		Labels:      labels,
		Description: strings.TrimSpace(description),
	}, nil
}

// applyTemplateVars applies Go template variables to a string.
func applyTemplateVars(text string, vars map[string]string) (string, error) {
	if text == "" {
		return "", nil
	}

	tmpl, err := template.New("").Parse(text)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ParseVarFlags parses --var flags in "key=value" format.
func ParseVarFlags(flags []string) (map[string]string, error) {
	vars := make(map[string]string)
	for _, flag := range flags {
		idx := strings.Index(flag, "=")
		if idx == -1 {
			return nil, fmt.Errorf("invalid variable format %q (expected key=value)", flag)
		}
		key := flag[:idx]
		value := flag[idx+1:]
		if key == "" {
			return nil, fmt.Errorf("variable key cannot be empty in %q", flag)
		}
		vars[key] = value
	}
	return vars, nil
}
