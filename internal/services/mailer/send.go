package mailer

import (
	"context"
	"fmt"
	"time"

	"github.com/wneessen/go-mail"
)

// Send delivers a message to one or more recipients.
func Send(ctx context.Context, cfg SMTPConfig, to []string, subject, textBody, htmlBody string) error {
	if !cfg.Ready() {
		return fmt.Errorf("smtp is not configured")
	}
	if len(to) == 0 {
		return fmt.Errorf("no recipients")
	}

	msg := mail.NewMsg()
	if err := msg.From(cfg.From); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}
	if err := msg.To(to...); err != nil {
		return fmt.Errorf("invalid recipients: %w", err)
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, textBody)
	msg.AddAlternativeString(mail.TypeTextHTML, htmlBody)

	timeout := cfg.Timeout()
	client, err := newClient(cfg.Host, cfg.Port, cfg.Username, cfg.Password, timeout)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()

	sendCtx, cancel := context.WithTimeout(ctx, timeout+10*time.Second)
	defer cancel()
	if err := client.DialAndSendWithContext(sendCtx, msg); err != nil {
		return fmt.Errorf("send mail: %w", err)
	}
	return nil
}
