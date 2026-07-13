// Package translate implements AI-powered content translation via configured providers.
package translate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/piplos/piplos.media/internal/config"
	"github.com/piplos/piplos.media/internal/repository"
	"github.com/piplos/piplos.media/internal/services/ai"
)

// Language display names for the translation prompt.
var languageNames = map[string]string{
	"en": "English", "ru": "Russian", "be": "Belarusian", "uk": "Ukrainian",
	"pl": "Polish", "de": "German", "fr": "French", "es": "Spanish",
	"it": "Italian", "pt": "Portuguese", "zh": "Chinese", "ja": "Japanese",
}

const defaultPrompt = `You are a professional translator for a software development agency website. Translate the values of the given JSON object into {target_language}. Keep the JSON keys unchanged, preserve technology names, brand names, numbers and formatting. Return only a valid JSON object.`

// LanguageDisplayName returns a human-readable language name for a code.
func LanguageDisplayName(code string) string {
	if name, ok := languageNames[strings.ToLower(code)]; ok {
		return name
	}
	return code
}

// Service translates content fields with an LLM.
type Service struct {
	repo *repository.Repository
}

// New creates a translate Service.
func New(repo *repository.Repository) *Service {
	return &Service{repo: repo}
}

// TranslateFields translates a map of field->text into targetLang, preserving keys.
func (s *Service) TranslateFields(ctx context.Context, fields map[string]string, targetLang string) (map[string]string, error) {
	client, task, err := s.resolveClient(ctx)
	if err != nil {
		return nil, err
	}

	systemPrompt := strings.ReplaceAll(task.Prompt, "{target_language}", LanguageDisplayName(targetLang))
	if strings.TrimSpace(systemPrompt) == "" {
		systemPrompt = strings.ReplaceAll(defaultPrompt, "{target_language}", LanguageDisplayName(targetLang))
	}

	sourceJSON, err := json.Marshal(fields)
	if err != nil {
		return nil, fmt.Errorf("marshal source fields: %w", err)
	}
	userPrompt := fmt.Sprintf("Input JSON to translate:\n%s", string(sourceJSON))

	content, err := client.ChatJSON(ctx, systemPrompt, userPrompt, 0.3)
	if err != nil {
		return nil, err
	}

	out := map[string]string{}
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		return nil, fmt.Errorf("unmarshal translated JSON: %w", err)
	}
	return out, nil
}

func (s *Service) resolveClient(ctx context.Context) (ai.Client, ai.TaskSettings, error) {
	rawTask, err := s.repo.GetDecryptedValue(ctx, config.KeyAITranslation)
	if err != nil {
		return nil, ai.TaskSettings{}, err
	}
	var task ai.TaskSettings
	if rawTask != "" {
		if err := json.Unmarshal([]byte(rawTask), &task); err != nil {
			return nil, ai.TaskSettings{}, fmt.Errorf("parse AI_TRANSLATION: %w", err)
		}
	}
	provider := strings.TrimSpace(strings.ToLower(task.Provider))
	model := strings.TrimSpace(task.Model)
	if provider == "" || model == "" {
		return nil, task, fmt.Errorf("AI translation provider/model is not configured")
	}

	providerKey := ai.ProviderSettingKey(provider)
	rawProvider, err := s.repo.GetDecryptedValue(ctx, providerKey)
	if err != nil {
		return nil, task, err
	}
	var ps ai.ProviderSettings
	if rawProvider != "" {
		if err := json.Unmarshal([]byte(rawProvider), &ps); err != nil {
			return nil, task, fmt.Errorf("parse %s settings: %w", providerKey, err)
		}
	}
	if !ps.Enable || ps.APIKey == "" {
		return nil, task, fmt.Errorf("provider %q is disabled or has no API key", provider)
	}

	timeout := time.Duration(ps.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return ai.NewClient(provider, ps.APIKey, model, timeout), task, nil
}

// TestTranslation runs a sample translation for admin testing.
func (s *Service) TestTranslation(ctx context.Context) (map[string]string, string, error) {
	fields := map[string]string{
		"title":       "Analytics Dashboard",
		"description": "Real-time metrics for operations teams.",
	}
	translated, err := s.TranslateFields(ctx, fields, "ru")
	if err != nil {
		return nil, "", err
	}
	userPrompt := fmt.Sprintf("Input JSON to translate:\n%s", mustJSON(fields))
	return translated, userPrompt, nil
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
