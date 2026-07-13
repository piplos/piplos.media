// Package mailer wraps SMTP connectivity for the settings test endpoint.
package mailer

import (
	"context"
	"errors"
	"time"

	"github.com/wneessen/go-mail"
)

// TestConnection dials the SMTP server and authenticates. No mail is sent.
func TestConnection(host string, port int, username, password string) error {
	timeout := 10 * time.Second
	client, err := newClient(host, port, username, password, timeout)
	if err != nil {
		return err
	}
	defer func() { _ = client.Close() }()
	ctx, cancel := context.WithTimeoutCause(context.Background(), timeout+5*time.Second, errors.New("smtp dial timeout"))
	defer cancel()
	return client.DialWithContext(ctx)
}

func newClient(host string, port int, username, password string, timeout time.Duration) (*mail.Client, error) {
	opts := []mail.Option{mail.WithPort(port), mail.WithTimeout(timeout)}
	if port == 465 {
		opts = append(opts, mail.WithSSLPort(true))
	} else {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSMandatory))
	}
	if username != "" || password != "" {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(username), mail.WithPassword(password))
	}
	return mail.NewClient(host, opts...)
}
