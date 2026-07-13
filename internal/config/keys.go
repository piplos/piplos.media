package config

// Composite setting keys (DB settings table).
const (
	KeyGemini     = "GEMINI"
	KeyGrok       = "GROK"
	KeyOpenAI     = "OPENAI"
	KeyOpenRouter = "OPENROUTER"
	KeySMTP       = "SMTP"

	KeyAITranslation = "AI_TRANSLATION"
)

// AllowedSettingKeys lists keys editable via admin API.
var AllowedSettingKeys = map[string]bool{
	KeyGemini: true, KeyGrok: true, KeyOpenAI: true, KeyOpenRouter: true,
	KeySMTP: true, KeyAITranslation: true,
}

// SensitiveFields maps composite keys to secret JSON field names.
var SensitiveFields = map[string][]string{
	KeyGemini:     {"apiKey"},
	KeyGrok:       {"apiKey"},
	KeyOpenAI:     {"apiKey"},
	KeyOpenRouter: {"apiKey"},
	KeySMTP:       {"username", "password"},
}

// ProviderByKey maps composite provider key to provider slug.
var ProviderByKey = map[string]string{
	KeyGemini:     "gemini",
	KeyGrok:       "grok",
	KeyOpenAI:     "openai",
	KeyOpenRouter: "openrouter",
}
