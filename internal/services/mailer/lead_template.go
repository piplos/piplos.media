package mailer

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/models"
)

// LeadEmail is a rendered admin notification for a new lead.
type LeadEmail struct {
	Subject  string
	TextBody string
	HTMLBody string
}

// LeadTemplate is the admin notification template stored in settings (see migration 005).
type LeadTemplate struct {
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// Ready reports whether the template is configured (non-empty body).
func (t LeadTemplate) Ready() bool {
	return strings.TrimSpace(t.Body) != ""
}

// LoadLeadTemplate reads the lead email template from settings.
// Missing or malformed values yield an empty (not ready) template.
func LoadLeadTemplate(ctx context.Context, repo smtpLoader) LeadTemplate {
	raw, err := repo.GetDecryptedValue(ctx, config.KeyLeadEmailTemplate)
	if err != nil || raw == "" {
		return LeadTemplate{}
	}
	var tpl LeadTemplate
	if json.Unmarshal([]byte(raw), &tpl) != nil {
		return LeadTemplate{}
	}
	return tpl
}

// LeadTemplateVars returns placeholder values for a lead ({{var}} -> value).
// Option-подобные поля (types, timeline, stage, how_found) подставляются
// в человекочитаемом виде на языке заявки.
func LeadTemplateVars(lead *models.Lead, adminURL string) map[string]string {
	lang := normalizeLang(lead.Lang)
	empty := localeFor(lang).empty
	fullName := strings.TrimSpace(strings.TrimSpace(lead.FirstName) + " " + strings.TrimSpace(lead.LastName))
	return map[string]string{
		"id":           lead.ID,
		"display_name": displayName(lead),
		"project_name": orEmpty(lead.ProjectName, empty),
		"types": joinTranslated(lead.Types, func(id string) string {
			return lookupOption(typeLabels, lang, id)
		}, empty),
		"description": orEmpty(lead.Description, empty),
		"stack":       orEmpty(lead.Stack, empty),
		"references":  orEmpty(lead.ReferenceURLs, empty),
		"budget":      formatBudget(lead.Budget, lead.Currency, empty),
		"timeline":    orEmpty(lookupOption(timelineLabels, lang, lead.Timeline), empty),
		"stage":       orEmpty(lookupOption(stageLabels, lang, lead.Stage), empty),
		"first_name":  orEmpty(lead.FirstName, empty),
		"last_name":   orEmpty(lead.LastName, empty),
		"name":        orEmpty(fullName, empty),
		"email":       lead.Email,
		"company":     orEmpty(lead.Company, empty),
		"phone":       orEmpty(lead.Phone, empty),
		"how_found":   orEmpty(lookupOption(howFoundLabels, lang, lead.HowFound), empty),
		"notes":       orEmpty(lead.Notes, empty),
		"lang":        strings.ToUpper(lang),
		"created_at":  formatLeadTime(lead.CreatedAt, lang),
		"lead_url":    leadAdminURL(adminURL, lead.ID),
	}
}

// RenderLeadEmail renders the template from settings, replacing {{var}} placeholders.
// The body is Markdown: the plain-text part keeps the raw text, the HTML part
// is rendered via goldmark (newlines become <br>).
func RenderLeadEmail(tpl LeadTemplate, lead *models.Lead, adminURL string) LeadEmail {
	vars := LeadTemplateVars(lead, adminURL)
	replace := func(s string) string {
		for key, value := range vars {
			s = strings.ReplaceAll(s, "{{"+key+"}}", value)
		}
		return s
	}

	subject := strings.TrimSpace(replace(tpl.Subject))
	if subject == "" {
		subject = displayName(lead)
	}
	textBody := replace(tpl.Body)

	htmlBody := strings.Builder{}
	htmlBody.WriteString("<!DOCTYPE html><html><body style=\"font-family:sans-serif;line-height:1.5;color:#111;\">")
	htmlBody.WriteString(markdownToHTML(textBody))
	htmlBody.WriteString("</body></html>")

	return LeadEmail{
		Subject:  subject,
		TextBody: textBody,
		HTMLBody: htmlBody.String(),
	}
}
