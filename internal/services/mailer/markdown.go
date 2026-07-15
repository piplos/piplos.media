package mailer

import (
	"bytes"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

// emailMarkdown renders the email template body. Hard wraps keep newlines as <br>;
// raw HTML is escaped because lead data is substituted into the template before rendering.
var emailMarkdown = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(html.WithHardWraps()),
)

// markdownToHTML converts Markdown to an HTML fragment (empty input -> "").
func markdownToHTML(src string) string {
	trimmed := strings.TrimSpace(src)
	if trimmed == "" {
		return ""
	}
	var buf bytes.Buffer
	if err := emailMarkdown.Convert([]byte(trimmed), &buf); err != nil {
		return trimmed
	}
	return strings.TrimSpace(buf.String())
}
