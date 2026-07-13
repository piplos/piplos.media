package mailer

import (
	"strings"
	"testing"
	"time"

	"github.com/piplos-media/site/internal/models"
)

func TestBuildLeadEmail_Russian(t *testing.T) {
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

	email := BuildLeadEmail(lead, "http://localhost:5174")

	if !strings.Contains(email.Subject, "Новая заявка") {
		t.Fatalf("subject: %q", email.Subject)
	}
	for _, want := range []string{
		"Веб-приложение", "Бэкенд / API", "1–3 месяца", "Только идея", "Рекомендация",
		"http://localhost:5174/leads/abc-123", "Открыть в админке",
	} {
		if !strings.Contains(email.TextBody, want) && !strings.Contains(email.HTMLBody, want) {
			t.Fatalf("missing %q in email", want)
		}
	}
}

func TestBuildLeadEmail_English(t *testing.T) {
	lead := &models.Lead{
		Types:       []string{"mobile"},
		ProjectName: "Delivery App",
		Timeline:    "flexible",
		Stage:       "mvp",
		FirstName:   "Alex",
		Email:       "alex@example.com",
		Lang:        "en",
	}

	email := BuildLeadEmail(lead, "")

	if !strings.Contains(email.Subject, "New project request") {
		t.Fatalf("subject: %q", email.Subject)
	}
	for _, want := range []string{"Mobile App", "Flexible", "Existing MVP"} {
		if !strings.Contains(email.TextBody, want) {
			t.Fatalf("missing %q", want)
		}
	}
	if strings.Contains(email.TextBody, "Open in admin") {
		t.Fatal("expected no admin link when admin URL is empty")
	}
}
