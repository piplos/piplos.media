// Package models defines domain entities shared by repository and handlers.
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// UserRole is the access role of an admin panel user.
type UserRole string

const (
	RoleAdmin   UserRole = "admin"
	RoleManager UserRole = "manager"
)

// User is an admin panel account.
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         UserRole  `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Language is a system content language.
type Language struct {
	Code      string `json:"code"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
}

// Translations maps language code -> field -> value.
type Translations map[string]map[string]string

// Value returns translations as raw JSON for storage.
func (t Translations) JSON() ([]byte, error) {
	if t == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(t)
}

// Project is a portfolio case study.
type Project struct {
	ID           string       `json:"id"`
	Slug         string       `json:"slug"`
	Category     string       `json:"category"`
	Categories   []string     `json:"categories"`
	Tags         []string     `json:"tags"`
	Year         int          `json:"year"`
	Featured     bool         `json:"featured"`
	Published    bool         `json:"published"`
	SortOrder    int          `json:"sort_order"`
	Image        string       `json:"image"`
	Translations Translations `json:"translations"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// Service is an offered service (web, mobile, backend, ...).
type Service struct {
	ID           string       `json:"id"`
	Slug         string       `json:"slug"`
	Icon         string       `json:"icon"`
	Tags         []string     `json:"tags"`
	Published    bool         `json:"published"`
	SortOrder    int          `json:"sort_order"`
	Translations Translations `json:"translations"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// StackItem is a technology in the tech stack.
type StackItem struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Label     string    `json:"label"`
	GroupID   string    `json:"group_id"`
	Published bool      `json:"published"`
	SortOrder int       `json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SEOPage stores per-path meta tags with translations.
type SEOPage struct {
	ID           string       `json:"id"`
	Path         string       `json:"path"`
	Translations Translations `json:"translations"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// LegalSection is a titled block inside a legal document.
type LegalSection struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// LegalLocale is legal document content for one language.
type LegalLocale struct {
	Label       string         `json:"label"`
	Title       string         `json:"title"`
	LastUpdated string         `json:"last_updated"`
	Sections    []LegalSection `json:"sections"`
}

// LegalTranslations maps language code to legal content.
type LegalTranslations map[string]LegalLocale

// LegalPage is a legal document (privacy policy, terms, cookies).
type LegalPage struct {
	ID           string            `json:"id"`
	Slug         string            `json:"slug"`
	Path         string            `json:"path"`
	SortOrder    int               `json:"sort_order"`
	Translations LegalTranslations `json:"translations"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
}

// LegalSlugs are fixed document identifiers.
var LegalSlugs = []string{"privacy", "terms", "cookies"}

// IsLegalPath reports whether path is reserved for legal documents (no standalone SEO).
func IsLegalPath(path string) bool {
	for _, slug := range LegalSlugs {
		if path == "/legal/"+slug {
			return true
		}
	}
	return false
}

// LeadStatus is the processing status of a site request.
type LeadStatus string

const (
	LeadNew        LeadStatus = "new"
	LeadInProgress LeadStatus = "in_progress"
	LeadDone       LeadStatus = "done"
	LeadSpam       LeadStatus = "spam"
)

// Lead is a request submitted from the site order form.
type Lead struct {
	ID            string     `json:"id"`
	Types         []string   `json:"types"`
	ProjectName   string     `json:"project_name"`
	Description   string     `json:"description"`
	Stack         string     `json:"stack"`
	ReferenceURLs string     `json:"references"`
	Budget        int        `json:"budget"`
	Currency      string     `json:"currency"`
	Timeline      string     `json:"timeline"`
	Stage         string     `json:"stage"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Email         string     `json:"email"`
	Company       string     `json:"company"`
	Phone         string     `json:"phone"`
	HowFound      string     `json:"how_found"`
	Notes         string     `json:"notes"`
	Lang          string     `json:"lang"`
	Status        LeadStatus `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Setting is a key/value configuration entry.
type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AIProviderModel is a row in ai_provider_models.
type AIProviderModel struct {
	ID          uuid.UUID `json:"id"`
	Provider    string    `json:"provider"`
	ModelID     string    `json:"model_id"`
	DisplayName string    `json:"display_name"`
	Enabled     bool      `json:"enabled"`
}
