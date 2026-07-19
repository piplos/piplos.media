package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/piplos/piplos.media/internal/config"
	apperrors "github.com/piplos/piplos.media/internal/errors"
	"github.com/piplos/piplos.media/internal/models"
	"github.com/piplos/piplos.media/internal/repository"
	"github.com/piplos/piplos.media/internal/services/ai"
	"github.com/piplos/piplos.media/internal/services/mailer"
	"github.com/piplos/piplos.media/internal/services/translate"
	"github.com/piplos/piplos.media/internal/storage"
	"github.com/piplos/piplos.media/internal/utils"
)

const maskedValue = "****"

// SettingsHandler manages settings, languages and AI translation.
type SettingsHandler struct {
	repo      *repository.Repository
	translate *translate.Service
}

// NewSettingsHandler creates a SettingsHandler.
func NewSettingsHandler(repo *repository.Repository, tr *translate.Service) *SettingsHandler {
	return &SettingsHandler{repo: repo, translate: tr}
}

func isAllowedSettingKey(key string) bool {
	return config.AllowedSettingKeys[strings.ToUpper(key)]
}

// maskSetting hides sensitive JSON fields in a setting value.
func maskSetting(s *models.Setting) {
	if fields := config.SensitiveFields[s.Key]; len(fields) > 0 {
		s.Value = utils.MaskJSONFields(s.Value, fields, maskedValue)
	}
}

// ListSettings returns all settings; sensitive fields are masked.
func (h *SettingsHandler) ListSettings(c fiber.Ctx) error {
	items, err := h.repo.ListSettings(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list settings")
	}
	for i := range items {
		maskSetting(&items[i])
	}
	return c.JSON(fiber.Map{"settings": items})
}

// GetSetting returns one setting by key (sensitive fields masked).
func (h *SettingsHandler) GetSetting(c fiber.Ctx) error {
	key := strings.ToUpper(c.Params("key"))
	if !isAllowedSettingKey(key) {
		return apperrors.ErrInvalidRequest("unknown setting key")
	}
	raw, err := h.repo.GetSetting(c.Context(), key)
	if err != nil {
		return apperrors.ErrInternal("failed to get setting")
	}
	if raw == "" {
		return apperrors.ErrNotFound("setting not found")
	}
	s := models.Setting{Key: key, Value: raw}
	maskSetting(&s)
	return c.JSON(fiber.Map{"key": s.Key, "value": s.Value})
}

type settingRequest struct {
	Value string `json:"value"`
}

// UpdateSetting saves a composite setting. Sensitive fields are encrypted at rest;
// masked ("****") sensitive fields keep their stored values.
func (h *SettingsHandler) UpdateSetting(c fiber.Ctx) error {
	key := strings.ToUpper(c.Params("key"))
	if !isAllowedSettingKey(key) {
		return apperrors.ErrInvalidRequest("unknown setting key")
	}
	var req settingRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(req.Value), &obj); err != nil {
		return apperrors.ErrInvalidRequest("value must be a JSON object")
	}
	if err := h.repo.SetCompositeSetting(c.Context(), key, req.Value, config.SensitiveFields[key]); err != nil {
		return apperrors.ErrInternal("failed to update setting")
	}
	return c.JSON(fiber.Map{"ok": true})
}

type testSettingRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type smtpSettings struct {
	Host           string `json:"host"`
	Port           int    `json:"port"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	From           string `json:"from"`
	TimeoutSeconds int    `json:"timeout_seconds"`
}

func extractAPIKeyFromProviderJSON(jsonStr string) string {
	var ps ai.ProviderSettings
	if json.Unmarshal([]byte(jsonStr), &ps) != nil {
		return ""
	}
	return ps.APIKey
}

func (h *SettingsHandler) firstEnabledModel(ctx context.Context, provider string) (string, error) {
	models, err := h.repo.ListEnabledAIProviderModels(ctx, provider)
	if err != nil {
		return "", err
	}
	if len(models) == 0 {
		return "", fmt.Errorf("no enabled models for provider")
	}
	return models[0].ModelID, nil
}

func (h *SettingsHandler) testAIProviderKey(ctx context.Context, provider, apiKey string) error {
	model, err := h.firstEnabledModel(ctx, provider)
	if err != nil {
		return fmt.Errorf("no enabled models for provider")
	}
	timeout := 15 * time.Second
	client := ai.NewClient(provider, apiKey, model, timeout)
	return client.TestAPIKey(ctx)
}

// testS3 verifies S3 credentials and bucket access. Masked secrets are
// substituted with stored values, so the saved config can be re-tested.
func (h *SettingsHandler) testS3(ctx context.Context, rawValue string) error {
	var cfg storage.S3Config
	if err := json.Unmarshal([]byte(rawValue), &cfg); err != nil {
		return fmt.Errorf("value must be a JSON object")
	}
	if cfg.AccessKeyID == maskedValue || cfg.SecretAccessKey == maskedValue {
		stored, err := h.repo.GetDecryptedValue(ctx, config.KeyS3)
		if err != nil {
			return fmt.Errorf("failed to load stored S3 settings")
		}
		var storedCfg storage.S3Config
		_ = json.Unmarshal([]byte(stored), &storedCfg)
		if cfg.AccessKeyID == maskedValue {
			cfg.AccessKeyID = storedCfg.AccessKeyID
		}
		if cfg.SecretAccessKey == maskedValue {
			cfg.SecretAccessKey = storedCfg.SecretAccessKey
		}
	}
	client, err := storage.NewS3(cfg)
	if err != nil {
		return err
	}
	return client.TestConnection(ctx)
}

// TestSetting checks a setting without saving it. Supported: SMTP, AI providers, S3.
func (h *SettingsHandler) TestSetting(c fiber.Ctx) error {
	var req testSettingRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	key := strings.ToUpper(req.Key)

	if key == config.KeyS3 {
		if err := h.testS3(c.Context(), strings.TrimSpace(req.Value)); err != nil {
			return apperrors.ErrInvalidRequest(err.Error())
		}
		return c.JSON(fiber.Map{"ok": true})
	}

	if provider, ok := config.ProviderByKey[key]; ok {
		apiKey := extractAPIKeyFromProviderJSON(strings.TrimSpace(req.Value))
		if apiKey == "" || apiKey == maskedValue {
			stored, err := h.repo.GetDecryptedValue(c.Context(), key)
			if err != nil {
				return apperrors.ErrInternal("failed to load stored provider settings")
			}
			apiKey = extractAPIKeyFromProviderJSON(stored)
		}
		if apiKey == "" {
			return apperrors.ErrInvalidRequest("API key is not configured")
		}
		if err := h.testAIProviderKey(c.Context(), provider, apiKey); err != nil {
			return apperrors.ErrInvalidRequest(err.Error())
		}
		return c.JSON(fiber.Map{"ok": true})
	}

	if key != config.KeySMTP {
		return apperrors.ErrInvalidRequest("unsupported test key")
	}

	var cfg smtpSettings
	if err := json.Unmarshal([]byte(req.Value), &cfg); err != nil {
		return apperrors.ErrInvalidRequest("value must be a JSON object")
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.Port < 1 || cfg.Port > 65535 {
		return apperrors.ErrInvalidRequest("invalid SMTP port")
	}
	if cfg.Username == maskedValue || cfg.Password == maskedValue {
		stored, err := h.repo.GetDecryptedValue(c.Context(), config.KeySMTP)
		if err != nil {
			return apperrors.ErrInternal("failed to load stored SMTP settings")
		}
		var storedCfg smtpSettings
		_ = json.Unmarshal([]byte(stored), &storedCfg)
		if cfg.Username == maskedValue {
			cfg.Username = storedCfg.Username
		}
		if cfg.Password == maskedValue {
			cfg.Password = storedCfg.Password
		}
	}
	if cfg.Host == "" {
		return apperrors.ErrInvalidRequest("SMTP host is required")
	}
	if err := mailer.TestConnection(cfg.Host, cfg.Port, cfg.Username, cfg.Password); err != nil {
		return apperrors.ErrInvalidRequest("SMTP connection failed: " + err.Error())
	}
	return c.JSON(fiber.Map{"ok": true})
}

// TestTranslation runs a sample content translation using current AI_TRANSLATION settings.
func (h *SettingsHandler) TestTranslation(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 100*time.Second)
	defer cancel()
	translated, userPrompt, err := h.translate.TestTranslation(ctx)
	if err != nil {
		return apperrors.New(apperrors.CodeServiceError, "translation test failed: "+err.Error())
	}
	return c.JSON(fiber.Map{
		"response": map[string]any{
			"content":     translated,
			"user_prompt": userPrompt,
		},
	})
}

// ---------- Languages ----------

// ListLanguages returns configured content languages.
func (h *SettingsHandler) ListLanguages(c fiber.Ctx) error {
	langs, err := h.repo.ListLanguages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to list languages")
	}
	return c.JSON(fiber.Map{"languages": langs})
}

type languageRequest struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
}

// UpsertLanguage creates or updates a language.
func (h *SettingsHandler) UpsertLanguage(c fiber.Ctx) error {
	var req languageRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	req.Code = strings.ToLower(strings.TrimSpace(req.Code))
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Code) < 2 || len(req.Code) > 5 || req.Name == "" {
		return apperrors.ErrInvalidRequest("code (2-5 chars) and name are required")
	}
	lang := models.Language{
		Code: req.Code, Name: req.Name, IsDefault: req.IsDefault,
		Enabled: req.Enabled, SortOrder: req.SortOrder,
	}
	if err := h.repo.UpsertLanguage(c.Context(), lang); err != nil {
		return apperrors.ErrInternal("failed to save language")
	}
	return c.JSON(fiber.Map{"language": lang})
}

// DeleteLanguage removes a language (default language is protected).
func (h *SettingsHandler) DeleteLanguage(c fiber.Ctx) error {
	code := c.Params("code")
	langs, err := h.repo.ListLanguages(c.Context())
	if err != nil {
		return apperrors.ErrInternal("failed to delete language")
	}
	for _, l := range langs {
		if l.Code == code && l.IsDefault {
			return apperrors.ErrInvalidRequest("cannot delete the default language")
		}
	}
	if err := h.repo.DeleteLanguage(c.Context(), code); err != nil {
		return apperrors.ErrInternal("failed to delete language")
	}
	return c.JSON(fiber.Map{"ok": true})
}

// ---------- AI translation ----------

type translateRequest struct {
	Fields     map[string]string `json:"fields"`
	TargetLang string            `json:"target_lang"`
}

// Translate translates arbitrary content fields into target_lang via the configured AI provider.
func (h *SettingsHandler) Translate(c fiber.Ctx) error {
	var req translateRequest
	if err := c.Bind().Body(&req); err != nil {
		return apperrors.ErrInvalidRequest("invalid request body")
	}
	if len(req.Fields) == 0 || req.TargetLang == "" {
		return apperrors.ErrInvalidRequest("fields and target_lang are required")
	}

	translated, err := h.translate.TranslateFields(c.Context(), req.Fields, req.TargetLang)
	if err != nil {
		return apperrors.New(apperrors.CodeServiceError, "translation failed: "+err.Error())
	}
	return c.JSON(fiber.Map{"fields": translated, "target_lang": req.TargetLang})
}
