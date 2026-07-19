package config

// Composite setting keys (DB settings table).
const (
	KeyGemini     = "GEMINI"
	KeyGrok       = "GROK"
	KeyOpenAI     = "OPENAI"
	KeyOpenRouter = "OPENROUTER"
	KeySMTP       = "SMTP"

	KeyAITranslation     = "AI_TRANSLATION"
	KeyLeadEmailTemplate = "LEAD_EMAIL_TEMPLATE"

	// KeyBackup — расписание и параметры резервного копирования.
	// KeyS3 — общее подключение к S3-совместимому хранилищу (Cloudflare R2);
	// используется бекапами и может переиспользоваться другими фичами админки.
	KeyBackup = "BACKUP"
	KeyS3     = "S3"
)

// AllowedSettingKeys lists keys editable via admin API.
var AllowedSettingKeys = map[string]bool{
	KeyGemini: true, KeyGrok: true, KeyOpenAI: true, KeyOpenRouter: true,
	KeySMTP: true, KeyAITranslation: true, KeyLeadEmailTemplate: true,
	KeyBackup: true, KeyS3: true,
}

// SensitiveFields maps composite keys to secret JSON field names.
var SensitiveFields = map[string][]string{
	KeyGemini:     {"apiKey"},
	KeyGrok:       {"apiKey"},
	KeyOpenAI:     {"apiKey"},
	KeyOpenRouter: {"apiKey"},
	KeySMTP:       {"username", "password"},
	KeyS3:         {"access_key_id", "secret_access_key"},
}

// ProviderByKey maps composite provider key to provider slug.
var ProviderByKey = map[string]string{
	KeyGemini:     "gemini",
	KeyGrok:       "grok",
	KeyOpenAI:     "openai",
	KeyOpenRouter: "openrouter",
}
