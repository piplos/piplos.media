package mailer

import (
	"fmt"
	"strings"
	"time"

	"github.com/piplos/piplos.media/internal/models"
)

func displayName(lead *models.Lead) string {
	if name := strings.TrimSpace(lead.ProjectName); name != "" {
		return name
	}
	return strings.TrimSpace(strings.TrimSpace(lead.FirstName) + " " + strings.TrimSpace(lead.LastName))
}

func orEmpty(v, empty string) string {
	if strings.TrimSpace(v) == "" {
		return empty
	}
	return v
}

func joinTranslated(items []string, translate func(string) string, empty string) string {
	if len(items) == 0 {
		return empty
	}
	out := make([]string, len(items))
	for i, id := range items {
		out[i] = translate(id)
	}
	return strings.Join(out, ", ")
}

func formatBudget(budget int, currency, empty string) string {
	if budget <= 0 {
		return empty
	}
	cur := strings.TrimSpace(currency)
	if cur == "" {
		cur = "USD"
	}
	return fmt.Sprintf("%d %s", budget, cur)
}

func formatLeadTime(t time.Time, lang string) string {
	if t.IsZero() {
		return localeFor(lang).empty
	}
	if lang == "ru" {
		return t.Format("02.01.2006 15:04")
	}
	return t.Format("Jan 2, 2006 3:04 PM")
}

func leadAdminURL(adminURL, leadID string) string {
	base := strings.TrimRight(strings.TrimSpace(adminURL), "/")
	if base == "" || leadID == "" {
		return ""
	}
	return base + "/leads/" + leadID
}
