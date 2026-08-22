package mailer

import (
	"strings"
	"testing"

	"github.com/piplos/piplos.media/internal/models"
)

// Значение из лида, похожее на плейсхолдер, не должно раскрываться в другой
// поле (раньше подстановка шла по map в случайном порядке и была рекурсивной).
func TestRenderLeadEmailDoesNotReexpandPlaceholdersInValues(t *testing.T) {
	lead := &models.Lead{
		ID:          "lead-1",
		ProjectName: "Проект",
		Description: "текст с {{notes}} внутри",
		Notes:       "СЕКРЕТНЫЕ_ЗАМЕТКИ",
		Email:       "a@b.c",
	}
	tpl := LeadTemplate{Subject: "s", Body: "{{description}} | {{email}}"}

	email := RenderLeadEmail(tpl, lead, "")

	if !strings.Contains(email.TextBody, "текст с {{notes}} внутри") {
		t.Fatalf("value-looking placeholder must stay literal: %q", email.TextBody)
	}
	if strings.Contains(email.TextBody, "СЕКРЕТНЫЕ_ЗАМЕТКИ") {
		t.Fatalf("lead value re-expanded another field: %q", email.TextBody)
	}
	if !strings.Contains(email.TextBody, "a@b.c") {
		t.Fatalf("real placeholder must still be replaced: %q", email.TextBody)
	}
}
