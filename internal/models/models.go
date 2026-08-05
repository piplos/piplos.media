// Package models defines domain entities shared by repository and handlers.
package models

import (
	"encoding/json"
	"strings"
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
	ID           string   `json:"id"`
	Email        string   `json:"email"`
	PasswordHash string   `json:"-"`
	FullName     string   `json:"full_name"`
	Role         UserRole `json:"role"`
	IsActive     bool     `json:"is_active"`
	// NotifyLeads включает письма о новых заявках; меняется только администратором.
	NotifyLeads bool      `json:"notify_leads"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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

// ProjectLinkKind identifies where a project link points.
type ProjectLinkKind string

const (
	ProjectLinkWebsite    ProjectLinkKind = "website"
	ProjectLinkGooglePlay ProjectLinkKind = "google_play"
	ProjectLinkAppStore   ProjectLinkKind = "app_store"
)

// ProjectLink is an external URL shown on a case study page.
type ProjectLink struct {
	URL   string          `json:"url"`
	Label string          `json:"label"`
	Kind  ProjectLinkKind `json:"kind"`
}

// Valid reports whether k is a known ProjectLinkKind.
func (k ProjectLinkKind) Valid() bool {
	switch k {
	case ProjectLinkWebsite, ProjectLinkGooglePlay, ProjectLinkAppStore:
		return true
	default:
		return false
	}
}

// NonNilProjectLinks returns links, or an empty non-nil slice when links is nil.
func NonNilProjectLinks(links []ProjectLink) []ProjectLink {
	if links == nil {
		return []ProjectLink{}
	}
	return links
}

// InferProjectLinkKind derives kind from a URL host when kind is missing/unknown.
func InferProjectLinkKind(url string) ProjectLinkKind {
	switch {
	case strings.Contains(url, "play.google.com"):
		return ProjectLinkGooglePlay
	case strings.Contains(url, "apps.apple.com"):
		return ProjectLinkAppStore
	default:
		return ProjectLinkWebsite
	}
}

// DefaultProjectLinkLabel returns a store name or the raw URL when label is empty.
func DefaultProjectLinkLabel(url string, kind ProjectLinkKind) string {
	switch kind {
	case ProjectLinkGooglePlay:
		return "Google Play"
	case ProjectLinkAppStore:
		return "App Store"
	default:
		return url
	}
}

// NormalizeProjectLinks trims, dedupes by URL, and fills missing kind/label.
// Always returns a non-nil slice.
func NormalizeProjectLinks(links []ProjectLink) []ProjectLink {
	links = NonNilProjectLinks(links)
	out := make([]ProjectLink, 0, len(links))
	seen := make(map[string]struct{}, len(links))
	for _, link := range links {
		url := strings.TrimSpace(link.URL)
		if url == "" {
			continue
		}
		if _, ok := seen[url]; ok {
			continue
		}
		seen[url] = struct{}{}

		kind := link.Kind
		if !kind.Valid() {
			kind = InferProjectLinkKind(url)
		}
		label := strings.TrimSpace(link.Label)
		if label == "" {
			label = DefaultProjectLinkLabel(url, kind)
		}
		out = append(out, ProjectLink{URL: url, Label: label, Kind: kind})
	}
	return out
}

// Project is a portfolio case study.
// SortOrder is the position inside its service group; GlobalSortOrder is the
// cross-group position used by the public "all projects" listing.
type Project struct {
	ID              string        `json:"id"`
	Slug            string        `json:"slug"`
	Category        string        `json:"category"`
	Categories      []string      `json:"categories"`
	Tags            []string      `json:"tags"`
	Year            int           `json:"year"`
	Featured        bool          `json:"featured"`
	Published       bool          `json:"published"`
	SortOrder       int           `json:"sort_order"`
	GlobalSortOrder int           `json:"global_sort_order"`
	Image           string        `json:"image"`
	Links           []ProjectLink `json:"links"`
	Translations    Translations  `json:"translations"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
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
	Icon      string    `json:"icon"`
	IconAlt   string    `json:"icon_alt"`
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

// Page is a custom site page (published in the site "Articles" section).
// Unlike legal documents, pages are created and deleted in the admin panel.
type Page struct {
	ID           string       `json:"id"`
	Slug         string       `json:"slug"`
	Published    bool         `json:"published"`
	PublishAt    *time.Time   `json:"publish_at"`
	Image        string       `json:"image"`
	Tags         []string     `json:"tags"`
	Translations Translations `json:"translations"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

// IsLive reports whether the page is visible on the public site:
// published and not scheduled for a future date.
func (p *Page) IsLive(now time.Time) bool {
	if !p.Published {
		return false
	}
	return p.PublishAt == nil || !p.PublishAt.After(now)
}

// EnabledLanguageCodes returns codes of enabled content languages.
func EnabledLanguageCodes(langs []Language) []string {
	codes := make([]string, 0, len(langs))
	for _, l := range langs {
		if l.Enabled {
			codes = append(codes, l.Code)
		}
	}
	return codes
}

// IsLegalPath reports whether path is reserved for legal documents (no standalone SEO).
// langCodes — enabled language codes from the languages table.
func IsLegalPath(path string, langCodes []string) bool {
	for _, slug := range LegalSlugs {
		if path == "/legal/"+slug {
			return true
		}
		for _, lang := range langCodes {
			if path == "/"+lang+"/legal/"+slug {
				return true
			}
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
