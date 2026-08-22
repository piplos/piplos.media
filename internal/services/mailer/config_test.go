package mailer

import (
	"context"
	"testing"

	"github.com/piplos/piplos.media/internal/config"
)

type fakeSMTPSettings map[string]string

func (f fakeSMTPSettings) GetDecryptedValue(_ context.Context, key string) (string, error) {
	return f[key], nil
}

// Ненастроенный SMTP ("") — это нулевой конфиг, а не ошибка парсинга:
// NotifyNewLead должен тихо пропустить отправку через Ready()==false.
func TestLoadSMTPUnsetSettingYieldsZeroConfig(t *testing.T) {
	cfg, err := LoadSMTP(context.Background(), fakeSMTPSettings{})
	if err != nil {
		t.Fatalf("unset settings must not error, got %v", err)
	}
	if cfg.Ready() {
		t.Fatal("unset settings must not be ready")
	}
	if cfg.Port != 587 {
		t.Fatalf("default port: got %d, want 587", cfg.Port)
	}
}

func TestLoadSMTPParsesConfiguredValues(t *testing.T) {
	raw := `{"host":"smtp.test","port":465,"username":"u","password":"p","from":"noreply@test"}`
	cfg, err := LoadSMTP(context.Background(), fakeSMTPSettings{config.KeySMTP: raw})
	if err != nil {
		t.Fatalf("LoadSMTP: %v", err)
	}
	if !cfg.Ready() || cfg.Host != "smtp.test" || cfg.Port != 465 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadSMTPInvalidJSONStillErrors(t *testing.T) {
	if _, err := LoadSMTP(context.Background(), fakeSMTPSettings{config.KeySMTP: "{broken"}); err == nil {
		t.Fatal("expected parse error for malformed JSON")
	}
}
