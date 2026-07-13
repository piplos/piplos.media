package mailer

import "strings"

type emailLocale struct {
	subject     string
	intro       string
	viewInAdmin string
	empty       string
	labels      map[string]string
}

var emailLocales = map[string]emailLocale{
	"ru": {
		subject:     "Новая заявка с сайта",
		intro:       "На сайте отправлена новая заявка на проект.",
		viewInAdmin: "Открыть в админке",
		empty:       "—",
		labels: map[string]string{
			"project_name": "Название проекта",
			"types":        "Тип проекта",
			"description":  "Описание",
			"stack":        "Стек",
			"references":   "Референсы",
			"budget":       "Бюджет",
			"timeline":     "Сроки",
			"stage":        "Стадия",
			"name":         "Имя",
			"email":        "Email",
			"company":      "Компания",
			"phone":        "Телефон",
			"how_found":    "Как нашли",
			"notes":        "Примечания",
			"lang":         "Язык сайта",
			"created_at":   "Дата",
		},
	},
	"en": {
		subject:     "New project request",
		intro:       "A new project request was submitted on the website.",
		viewInAdmin: "Open in admin panel",
		empty:       "—",
		labels: map[string]string{
			"project_name": "Project name",
			"types":        "Project type",
			"description":  "Description",
			"stack":        "Tech stack",
			"references":   "References",
			"budget":       "Budget",
			"timeline":     "Timeline",
			"stage":        "Stage",
			"name":         "Name",
			"email":        "Email",
			"company":      "Company",
			"phone":        "Phone",
			"how_found":    "How they found us",
			"notes":        "Notes",
			"lang":         "Site language",
			"created_at":   "Submitted at",
		},
	},
}

var typeLabels = map[string]map[string]string{
	"ru": {
		"web": "Веб-приложение", "mobile": "Мобильное приложение", "backend": "Бэкенд / API",
		"data": "Данные и аналитика", "devops": "DevOps", "saas": "SaaS продукт", "consulting": "Консалтинг",
	},
	"en": {
		"web": "Web App", "mobile": "Mobile App", "backend": "Backend / API",
		"data": "Data & Analytics", "devops": "DevOps", "saas": "SaaS Product", "consulting": "Consulting",
	},
}

var stageLabels = map[string]map[string]string{
	"ru": {
		"idea": "Только идея", "design": "Есть дизайн", "mvp": "Существующий MVP", "legacy": "Миграция legacy",
	},
	"en": {
		"idea": "Just an Idea", "design": "Have Design", "mvp": "Existing MVP", "legacy": "Legacy Migration",
	},
}

var timelineLabels = map[string]map[string]string{
	"ru": {
		"1m": "Меньше 1 месяца", "1-3m": "1–3 месяца", "3-6m": "3–6 месяцев",
		"6-12m": "6–12 месяцев", "12m+": "12+ месяцев / постоянно", "flexible": "Гибкие",
	},
	"en": {
		"1m": "Less than 1 month", "1-3m": "1–3 months", "3-6m": "3–6 months",
		"6-12m": "6–12 months", "12m+": "12+ months / Ongoing", "flexible": "Flexible",
	},
}

var howFoundLabels = map[string]map[string]string{
	"ru": {
		"google": "Поиск Google", "referral": "Рекомендация", "social": "Социальные сети",
		"github": "GitHub", "other": "Другое",
	},
	"en": {
		"google": "Google Search", "referral": "Referral", "social": "Social Media",
		"github": "GitHub", "other": "Other",
	},
}

func normalizeLang(lang string) string {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang == "ru" {
		return "ru"
	}
	return "en"
}

func localeFor(lang string) emailLocale {
	if loc, ok := emailLocales[normalizeLang(lang)]; ok {
		return loc
	}
	return emailLocales["en"]
}

func labelFor(lang, key string) string {
	loc := localeFor(lang)
	if v, ok := loc.labels[key]; ok {
		return v
	}
	return key
}

func lookupOption(m map[string]map[string]string, lang, id string) string {
	if labels, ok := m[lang]; ok {
		if v, ok := labels[id]; ok {
			return v
		}
	}
	if labels, ok := m["en"]; ok {
		if v, ok := labels[id]; ok {
			return v
		}
	}
	return id
}
