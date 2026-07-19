package handlers

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/piplos/piplos.media/internal/models"
)

var markdownRenderer = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	// Raw HTML passes through: legacy content was saved as HTML by the old
	// rich-text editor. The site sanitizes output before rendering.
	goldmark.WithRendererOptions(html.WithUnsafe()),
)

// renderMarkdownFields returns a copy of translations where the listed fields
// are converted from Markdown to HTML (for public API responses).
func renderMarkdownFields(t models.Translations, fields ...string) models.Translations {
	if t == nil {
		return nil
	}
	out := make(models.Translations, len(t))
	for lang, values := range t {
		copied := make(map[string]string, len(values))
		for key, value := range values {
			copied[key] = value
		}
		for _, field := range fields {
			if value, ok := copied[field]; ok {
				copied[field] = markdownToHTML(value)
			}
		}
		out[lang] = copied
	}
	return out
}

// legacyHTMLRe detects content saved as HTML by the old rich-text editor: it
// always starts with a block-level tag. Markdown that begins with an inline
// tag (e.g. an <a class="btn-…"> inserted by the editor) is still rendered.
var legacyHTMLRe = regexp.MustCompile(`(?i)^<(p|div|h[1-6]|ul|ol|table|blockquote|figure|section|article|pre|img|br)[\s>/]`)

// markdownToHTML converts Markdown to HTML. Content that already starts with a
// block-level tag is treated as legacy HTML and returned as is.
func markdownToHTML(src string) string {
	trimmed := strings.TrimSpace(src)
	if trimmed == "" {
		return ""
	}
	if legacyHTMLRe.MatchString(trimmed) {
		return trimmed
	}
	var buf bytes.Buffer
	if err := markdownRenderer.Convert([]byte(trimmed), &buf); err != nil {
		return trimmed
	}
	return strings.TrimSpace(buf.String())
}

// renderLegalMarkdown returns a copy of legal translations where section bodies
// are converted from Markdown to HTML (for public API responses).
func renderLegalMarkdown(t models.LegalTranslations) models.LegalTranslations {
	if t == nil {
		return nil
	}
	out := make(models.LegalTranslations, len(t))
	for lang, locale := range t {
		sections := make([]models.LegalSection, len(locale.Sections))
		for i, section := range locale.Sections {
			sections[i] = models.LegalSection{
				Title: section.Title,
				Body:  markdownToHTML(section.Body),
			}
		}
		out[lang] = models.LegalLocale{
			Label:       locale.Label,
			Title:       locale.Title,
			LastUpdated: locale.LastUpdated,
			Sections:    sections,
		}
	}
	return out
}
