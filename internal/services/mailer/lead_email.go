package mailer

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/piplos/site/internal/models"
)

type leadField struct {
	label string
	value string
}

// LeadEmail is a rendered admin notification for a new lead.
type LeadEmail struct {
	Subject  string
	TextBody string
	HTMLBody string
}

// BuildLeadEmail renders a bilingual admin notification (lead language + English fallback labels in subject).
func BuildLeadEmail(lead *models.Lead, adminURL string) LeadEmail {
	lang := normalizeLang(lead.Lang)
	loc := localeFor(lang)

	subject := fmt.Sprintf("%s: %s", loc.subject, displayName(lead))
	fields := leadFields(lead, lang)

	text := strings.Builder{}
	text.WriteString(loc.intro)
	text.WriteByte('\n')
	text.WriteByte('\n')
	for _, f := range fields {
		text.WriteString(f.label)
		text.WriteString(": ")
		text.WriteString(f.value)
		text.WriteByte('\n')
	}
	if link := leadAdminURL(adminURL, lead.ID); link != "" {
		text.WriteByte('\n')
		text.WriteString(loc.viewInAdmin)
		text.WriteString(": ")
		text.WriteString(link)
		text.WriteByte('\n')
	}

	htmlBody := strings.Builder{}
	htmlBody.WriteString("<!DOCTYPE html><html><body style=\"font-family:sans-serif;line-height:1.5;color:#111;\">")
	htmlBody.WriteString("<p>")
	htmlBody.WriteString(html.EscapeString(loc.intro))
	htmlBody.WriteString("</p><table style=\"border-collapse:collapse;width:100%;max-width:640px;\">")
	for _, f := range fields {
		htmlBody.WriteString("<tr><td style=\"padding:6px 12px 6px 0;vertical-align:top;font-weight:600;white-space:nowrap;\">")
		htmlBody.WriteString(html.EscapeString(f.label))
		htmlBody.WriteString("</td><td style=\"padding:6px 0;vertical-align:top;\">")
		htmlBody.WriteString(strings.ReplaceAll(html.EscapeString(f.value), "\n", "<br>"))
		htmlBody.WriteString("</td></tr>")
	}
	htmlBody.WriteString("</table>")
	if link := leadAdminURL(adminURL, lead.ID); link != "" {
		htmlBody.WriteString("<p><a href=\"")
		htmlBody.WriteString(html.EscapeString(link))
		htmlBody.WriteString("\">")
		htmlBody.WriteString(html.EscapeString(loc.viewInAdmin))
		htmlBody.WriteString("</a></p>")
	}
	htmlBody.WriteString("</body></html>")

	return LeadEmail{
		Subject:  subject,
		TextBody: text.String(),
		HTMLBody: htmlBody.String(),
	}
}

func leadFields(lead *models.Lead, lang string) []leadField {
	loc := localeFor(lang)
	empty := loc.empty

	fullName := strings.TrimSpace(strings.TrimSpace(lead.FirstName) + " " + strings.TrimSpace(lead.LastName))
	fields := []leadField{
		{labelFor(lang, "project_name"), orEmpty(lead.ProjectName, empty)},
		{labelFor(lang, "types"), joinTranslated(lead.Types, func(id string) string {
			return lookupOption(typeLabels, lang, id)
		}, empty)},
		{labelFor(lang, "description"), orEmpty(lead.Description, empty)},
		{labelFor(lang, "stack"), orEmpty(lead.Stack, empty)},
		{labelFor(lang, "references"), orEmpty(lead.ReferenceURLs, empty)},
		{labelFor(lang, "budget"), formatBudget(lead.Budget, lead.Currency, empty)},
		{labelFor(lang, "timeline"), orEmpty(lookupOption(timelineLabels, lang, lead.Timeline), empty)},
		{labelFor(lang, "stage"), orEmpty(lookupOption(stageLabels, lang, lead.Stage), empty)},
		{labelFor(lang, "name"), orEmpty(fullName, empty)},
		{labelFor(lang, "email"), lead.Email},
		{labelFor(lang, "company"), orEmpty(lead.Company, empty)},
		{labelFor(lang, "phone"), orEmpty(lead.Phone, empty)},
		{labelFor(lang, "how_found"), orEmpty(lookupOption(howFoundLabels, lang, lead.HowFound), empty)},
		{labelFor(lang, "notes"), orEmpty(lead.Notes, empty)},
		{labelFor(lang, "lang"), strings.ToUpper(lang)},
		{labelFor(lang, "created_at"), formatLeadTime(lead.CreatedAt, lang)},
	}
	return fields
}

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
