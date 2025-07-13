package app

import (
	"context"
	"fmt"

	"github.com/p2p-b2b/mailer"
)

// initMailService initializes the mail service based on configuration
func (a *App) initMailService(ctx context.Context) error {
	var mailerService mailer.MailerService
	var err error

	switch a.configs.Mail.MailSender.Value {
	case "smtp":
		mailerService, err = mailer.NewMailerSMTP(mailer.MailerSMTPConf{
			SMTPHost: a.configs.Mail.SMTPHost.Value,
			SMTPPort: a.configs.Mail.SMTPPort.Value,
			Username: a.configs.Mail.SMTPUsername.Value,
			Password: a.configs.Mail.SMTPPassword.Value,
		})
		if err != nil {
			return fmt.Errorf("error creating SMTP mail service: %w", err)
		}

	case "mailgun":
		return fmt.Errorf("mailgun mailer not implemented yet")
	default:
		return fmt.Errorf("unknown mail sender type: %s", a.configs.Mail.MailSender.Value)
	}

	a.mailServer, err = mailer.NewMailService(&mailer.MailServiceConfig{
		Ctx:         ctx,
		WorkerCount: a.configs.Mail.MailWorkerCount.Value,
		Timeout:     a.configs.Mail.MailWorkerTimeout.Value,
		Mailer:      mailerService,
	})
	if err != nil {
		return fmt.Errorf("error creating mail service: %w", err)
	}

	return nil
}
