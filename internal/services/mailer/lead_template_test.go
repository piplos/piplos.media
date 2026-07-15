package mailer

import (
	"strings"
	"testing"
	"time"

	"github.com/piplos/piplos.media/internal/models"
)

// DefaultLeadTemplate matches migrations/005_add_lead_notifications.sql seed.
func DefaultLeadTemplate() LeadTemplate {
	return LeadTemplate{
		Subject: "{{name}} — новая заявка ({{project_name}})",
		Body: "## {{name}}\n" +
			"{{email}} · {{phone}}  \n" +
			"{{company}}\n\n" +
			"---\n\n" +
			"### Запрос\n\n" +
			"| | |\n" +
			"| --- | --- |\n" +
			"| **Проект** | {{project_name}} |\n" +
			"| **Тип** | {{types}} |\n" +
			"| **Бюджет** | {{budget}} |\n" +
			"| **Сроки** | {{timeline}} |\n" +
			"| **Стадия** | {{stage}} |\n\n" +
			"**Описание**  \n" +
			"{{description}}\n\n" +
			"**Стек:** {{stack}}  \n" +
			"**Референсы:** {{references}}\n\n" +
			"**Как нашли нас:** {{how_found}}  \n" +
			"**Примечания:** {{notes}}\n\n" +
			"_Заявка № {{id}} · {{created_at}} · {{lang}}_\n\n" +
			"[Открыть в админке →]({{lead_url}})",
	}
}

func TestRenderLeadEmail_ReplacesVariables(t *testing.T) {
	lead := &models.Lead{
		ID:          "abc-123",
		Types:       []string{"web"},
		ProjectName: "Аналитика",
		Budget:      5000,
		Currency:    "USD",
		FirstName:   "Иван",
		LastName:    "Петров",
		Email:       "ivan@example.com",
		Phone:       "+375 29 000 00 00",
		Lang:        "ru",
	}
	tpl := LeadTemplate{
		Subject: "Заявка: {{project_name}} от {{name}}",
		Body:    "Тип: {{types}}\nБюджет: {{budget}}\nEmail: {{email}}\nСсылка: {{lead_url}}",
	}

	email := RenderLeadEmail(tpl, lead, "http://localhost:5174")

	if email.Subject != "Заявка: Аналитика от Иван Петров" {
		t.Fatalf("subject: %q", email.Subject)
	}
	for _, want := range []string{
		"Тип: Веб-приложение", "Бюджет: 5000 USD", "Email: ivan@example.com",
		"Ссылка: http://localhost:5174/leads/abc-123",
	} {
		if !strings.Contains(email.TextBody, want) {
			t.Fatalf("missing %q in text body: %q", want, email.TextBody)
		}
	}
	if !strings.Contains(email.HTMLBody, "Веб-приложение") || !strings.Contains(email.HTMLBody, "<br>") {
		t.Fatalf("html body: %q", email.HTMLBody)
	}
}

func TestRenderLeadEmail_DefaultMigrationTemplate(t *testing.T) {
	lead := &models.Lead{
		ID:            "abc-123",
		Types:         []string{"web", "backend"},
		ProjectName:   "Аналитика",
		Description:   "Нужна платформа",
		Stack:         "Go, Svelte",
		ReferenceURLs: "https://example.com",
		Budget:        15000,
		Currency:      "BYN",
		Timeline:      "1-3m",
		Stage:         "idea",
		FirstName:     "Иван",
		LastName:      "Петров",
		Email:         "ivan@example.com",
		Company:       "ООО Тест",
		Phone:         "+375 29 000 00 00",
		HowFound:      "referral",
		Notes:         "NDA",
		Lang:          "ru",
		CreatedAt:     time.Date(2026, 7, 13, 15, 30, 0, 0, time.UTC),
	}

	email := RenderLeadEmail(DefaultLeadTemplate(), lead, "https://admin.piplos.media")

	if email.Subject != "Иван Петров — новая заявка (Аналитика)" {
		t.Fatalf("subject: %q", email.Subject)
	}
	for _, want := range []string{
		"## Иван Петров",
		"ivan@example.com · +375 29 000 00 00",
		"ООО Тест",
		"Веб-приложение", "Бэкенд / API", "1–3 месяца", "Только идея", "Рекомендация",
		"Заявка № abc-123",
		"https://admin.piplos.media/leads/abc-123",
	} {
		if !strings.Contains(email.TextBody, want) {
			t.Fatalf("missing %q in email body", want)
		}
	}
	for _, want := range []string{
		"<h2>Иван Петров</h2>",
		"<h3>Запрос</h3>",
		"<table>",
		"<strong>Проект</strong>",
		`<a href="https://admin.piplos.media/leads/abc-123">Открыть в админке →</a>`,
	} {
		if !strings.Contains(email.HTMLBody, want) {
			t.Fatalf("missing %q in html body: %q", want, email.HTMLBody)
		}
	}
}

func TestLeadTemplate_Ready(t *testing.T) {
	if (LeadTemplate{Subject: "s"}).Ready() {
		t.Fatal("template without body must not be ready")
	}
	if !(LeadTemplate{Body: "b"}).Ready() {
		t.Fatal("template with body must be ready")
	}
}
