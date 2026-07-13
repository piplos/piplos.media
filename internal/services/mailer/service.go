package mailer

import (
	"context"
	"strings"

	"github.com/rs/zerolog"

	"github.com/piplos/site/internal/models"
)

type userLister interface {
	ListUsers(ctx context.Context) ([]models.User, error)
}

// Service sends transactional email using SMTP settings from the database.
type Service struct {
	repo     smtpLoader
	users    userLister
	adminURL string
	log      zerolog.Logger
}

// NewService creates a mailer service.
func NewService(repo smtpLoader, users userLister, adminURL string, log zerolog.Logger) *Service {
	return &Service{repo: repo, users: users, adminURL: adminURL, log: log}
}

// NotifyNewLead emails active admins and managers about a submitted lead.
// Errors are logged; callers should not fail the HTTP request because of mail issues.
func (s *Service) NotifyNewLead(ctx context.Context, lead *models.Lead) {
	if lead == nil {
		return
	}

	cfg, err := LoadSMTP(ctx, s.repo)
	if err != nil {
		s.log.Warn().Err(err).Msg("lead notification: load smtp settings")
		return
	}
	if !cfg.Ready() {
		s.log.Debug().Msg("lead notification: smtp not configured, skipping")
		return
	}

	recipients, err := s.adminRecipients(ctx)
	if err != nil {
		s.log.Warn().Err(err).Msg("lead notification: list recipients")
		return
	}
	if len(recipients) == 0 {
		s.log.Warn().Msg("lead notification: no active admin recipients")
		return
	}

	email := BuildLeadEmail(lead, s.adminURL)
	if err := Send(ctx, cfg, recipients, email.Subject, email.TextBody, email.HTMLBody); err != nil {
		s.log.Warn().Err(err).Strs("to", recipients).Str("lead_id", lead.ID).Msg("lead notification: send failed")
		return
	}
	s.log.Info().Str("lead_id", lead.ID).Strs("to", recipients).Msg("lead notification sent")
}

func (s *Service) adminRecipients(ctx context.Context) ([]string, error) {
	users, err := s.users.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(users))
	out := make([]string, 0, len(users))
	for _, u := range users {
		if !u.IsActive {
			continue
		}
		if u.Role != models.RoleAdmin && u.Role != models.RoleManager {
			continue
		}
		email := strings.TrimSpace(strings.ToLower(u.Email))
		if email == "" || !strings.Contains(email, "@") {
			continue
		}
		if _, ok := seen[email]; ok {
			continue
		}
		seen[email] = struct{}{}
		out = append(out, u.Email)
	}
	return out, nil
}
