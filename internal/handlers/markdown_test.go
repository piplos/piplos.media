package handlers

import (
	"strings"
	"testing"

	"github.com/piplos/piplos.media/internal/models"
)

func TestMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "   ", ""},
		{"paragraph", "Просто текст", "<p>Просто текст</p>"},
		{"formatting", "**жирный** и *курсив*", "<p><strong>жирный</strong> и <em>курсив</em></p>"},
		{"heading and list", "## Заголовок\n\n- один\n- два", "<h2>Заголовок</h2>\n<ul>\n<li>один</li>\n<li>два</li>\n</ul>"},
		{"image", "![alt](https://example.com/a.png)", `<p><img src="https://example.com/a.png" alt="alt"></p>`},
		{"legacy html passthrough", "<p>Старый <strong>HTML</strong></p>", "<p>Старый <strong>HTML</strong></p>"},
		{
			"markdown starting with inline link",
			`<a href="https://example.com" class="btn-primary">Кнопка</a>` + "\n\n**дальше** markdown",
			"<p><a href=\"https://example.com\" class=\"btn-primary\">Кнопка</a></p>\n<p><strong>дальше</strong> markdown</p>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := markdownToHTML(tt.in); got != tt.want {
				t.Errorf("markdownToHTML(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRenderMarkdownFields(t *testing.T) {
	in := models.Translations{
		"ru": {"title": "# не трогаем", "body": "# Заголовок"},
		"en": {"title": "Title", "body": ""},
	}
	out := renderMarkdownFields(in, "body")

	if got := out["ru"]["body"]; !strings.Contains(got, "<h1>") {
		t.Errorf("body not rendered: %q", got)
	}
	if got := out["ru"]["title"]; got != "# не трогаем" {
		t.Errorf("title must stay raw, got %q", got)
	}
	// исходная карта не должна меняться
	if in["ru"]["body"] != "# Заголовок" {
		t.Errorf("source translations mutated: %q", in["ru"]["body"])
	}
	if renderMarkdownFields(nil, "body") != nil {
		t.Error("nil translations must stay nil")
	}
}

func TestRenderLegalMarkdown(t *testing.T) {
	in := models.LegalTranslations{
		"en": {
			Title: "Terms",
			Sections: []models.LegalSection{
				{Title: "1. General", Body: "**bold** text"},
			},
		},
	}
	out := renderLegalMarkdown(in)
	if got := out["en"].Sections[0].Body; !strings.Contains(got, "<strong>") {
		t.Errorf("section body not rendered: %q", got)
	}
	if in["en"].Sections[0].Body != "**bold** text" {
		t.Error("source legal translations mutated")
	}
	if renderLegalMarkdown(nil) != nil {
		t.Error("nil legal translations must stay nil")
	}
}
