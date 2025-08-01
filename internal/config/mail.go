package config

import (
	"errors"
	"net/mail"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

var (
	ErrMailSMTPHostOrAPIURLMustBeSet = errors.New("smtp host or api url must be set")
	ErrMailInvalidSMTPPort           = errors.New("invalid smtp port. Must be one of [" + ValidMailSMTPPorts + "]")
	ErrMailSMTPUsernameMustBeSet     = errors.New("smtp username must be set")
	ErrMailSMTPPasswordMustBeSet     = errors.New("smtp password must be set")
	ErrMailAPIURLInvalid             = errors.New("mail api url is invalid")
	ErrMailAPIKeyMustBeSet           = errors.New("mail api key must be set")
	ErrMailSenderNameMustBeSet       = errors.New("sender name must be set")
	ErrMailSenderAddressInvalid      = errors.New("sender address is invalid")
	ErrMailInvalidSender             = errors.New("invalid mail sender. Must be one of [" + ValidMailSender + "]")
	ErrMailInvalidWorkerCount        = errors.New("invalid mail worker count. Must be between [" + strconv.Itoa(ValidMailMinWorkerCount) + "] and [" + strconv.Itoa(ValidMailMaxWorkerCount) + "]")
	ErrMailInvalidWorkerTimeout      = errors.New("invalid mail worker timeout. Must be between [" + ValidMailMinWorkerTimeout.String() + "] and [" + ValidMailMaxWorkerTimeout.String() + "]")
)

const (
	ValidMailSMTPPorts        = "25|465|587|1025|2525"
	ValidMailSender           = "smtp|mailgun"
	ValidMailMaxWorkerCount   = 50
	ValidMailMinWorkerCount   = 1
	ValidMailMaxWorkerTimeout = 10 * time.Second
	ValidMailMinWorkerTimeout = 1 * time.Second

	DefaultMailSMTPHost      = ""
	DefaultMailSMTPUsername  = ""
	DefaultMailSMTPPassword  = ""
	DefaultMailSMTPPort      = 587
	DefaultMailSenderName    = "qu3ry me"
	DefaultMailSenderAddress = "no-reply@qu3ry.me"
	DefaultMailAPIEndpoint   = ""
	DefaultMailAPIKey        = ""
	DefaultMailWorkerCount   = 5
	DefaultMailWorkerTimeout = 5 * time.Second
	DefaultMailSender        = "smtp"
)

type MailConfig struct {
	// When smtp is used
	SMTPHost     Field[string]
	SMTPPort     Field[int]
	SMTPUsername Field[string]
	SMTPPassword Field[string]

	SenderName    Field[string]
	SenderAddress Field[string]

	// when mailgun or other service is used
	APIURL Field[string]
	APIKey Field[string]

	MailSender        Field[string]
	MailWorkerCount   Field[int]
	MailWorkerTimeout Field[time.Duration]
}

func NewMailConfig() *MailConfig {
	return &MailConfig{
		SMTPHost:          NewField("mail.smtp.host", "MAIL_SMTP_HOST", "SMTP Host", DefaultMailSMTPHost),
		SMTPPort:          NewField("mail.smtp.port", "MAIL_SMTP_PORT", "SMTP Port", DefaultMailSMTPPort),
		SMTPUsername:      NewField("mail.smtp.username", "MAIL_SMTP_USERNAME", "SMTP Username", DefaultMailSMTPUsername),
		SMTPPassword:      NewField("mail.smtp.password", "MAIL_SMTP_PASSWORD", "SMTP Password", DefaultMailSMTPPassword),
		SenderName:        NewField("mail.sender.name", "MAIL_SENDER_NAME", "Sender Name", DefaultMailSenderName),
		SenderAddress:     NewField("mail.sender.address", "MAIL_SENDER_ADDRESS", "Sender Address", DefaultMailSenderAddress),
		APIURL:            NewField("mail.api.url", "MAIL_API_URL", "Mail API URL", DefaultMailAPIEndpoint),
		APIKey:            NewField("mail.api.key", "MAIL_API_KEY", "Mail API Key", DefaultMailAPIKey),
		MailSender:        NewField("mail.sender", "MAIL_SENDER", "Mail Sender", DefaultMailSender),
		MailWorkerCount:   NewField("mail.worker.count", "MAIL_WORKER_COUNT", "Mail Worker Count", DefaultMailWorkerCount),
		MailWorkerTimeout: NewField("mail.worker.timeout", "MAIL_WORKER_TIMEOUT", "Mail Worker Timeout", DefaultMailWorkerTimeout),
	}
}

func (ref *MailConfig) ParseEnvVars() {
	ref.SMTPHost.Value = GetEnv(ref.SMTPHost.EnVarName, ref.SMTPHost.Value)
	ref.SMTPPort.Value = GetEnv(ref.SMTPPort.EnVarName, ref.SMTPPort.Value)
	ref.SMTPUsername.Value = GetEnv(ref.SMTPUsername.EnVarName, ref.SMTPUsername.Value)
	ref.SMTPPassword.Value = GetEnv(ref.SMTPPassword.EnVarName, ref.SMTPPassword.Value)
	ref.SenderName.Value = GetEnv(ref.SenderName.EnVarName, ref.SenderName.Value)
	ref.SenderAddress.Value = GetEnv(ref.SenderAddress.EnVarName, ref.SenderAddress.Value)
	ref.APIURL.Value = GetEnv(ref.APIURL.EnVarName, ref.APIURL.Value)
	ref.APIKey.Value = GetEnv(ref.APIKey.EnVarName, ref.APIKey.Value)
	ref.MailSender.Value = GetEnv(ref.MailSender.EnVarName, ref.MailSender.Value)
	ref.MailWorkerCount.Value = GetEnv(ref.MailWorkerCount.EnVarName, ref.MailWorkerCount.Value)
	ref.MailWorkerTimeout.Value = GetEnv(ref.MailWorkerTimeout.EnVarName, ref.MailWorkerTimeout.Value)
}

func (ref *MailConfig) Validate() error {
	if ref.SMTPHost.Value == "" && ref.APIURL.Value == "" {
		return ErrMailSMTPHostOrAPIURLMustBeSet
	}

	if ref.SMTPHost.Value != "" && !slices.Contains(strings.Split(ValidMailSMTPPorts, "|"), strconv.Itoa(ref.SMTPPort.Value)) {
		return ErrMailInvalidSMTPPort
	}

	if ref.MailSender.Value != "" && !slices.Contains(strings.Split(ValidMailSender, "|"), ref.MailSender.Value) {
		return ErrMailInvalidSender
	}

	if ref.MailSender.Value == "smtp" {
		if ref.SMTPHost.Value != "" && ref.SMTPUsername.Value == "" {
			return ErrMailSMTPUsernameMustBeSet
		}

		if ref.SMTPUsername.Value != "" && ref.SMTPPassword.Value == "" {
			return ErrMailSMTPPasswordMustBeSet
		}
	}

	if ref.MailSender.Value == "mailgun" {
		if ref.APIURL.Value != "" {
			if _, err := url.Parse(ref.APIURL.Value); err != nil {
				return ErrMailAPIURLInvalid
			}

			if ref.APIKey.Value == "" {
				return ErrMailAPIKeyMustBeSet
			}
		}
	}

	if ref.SenderName.Value == "" {
		return ErrMailSenderNameMustBeSet
	}

	if _, err := mail.ParseAddress(ref.SenderAddress.Value); err != nil {
		return ErrMailSenderAddressInvalid
	}

	if ref.MailWorkerCount.Value < ValidMailMinWorkerCount || ref.MailWorkerCount.Value > ValidMailMaxWorkerCount {
		return ErrMailInvalidWorkerCount
	}

	if ref.MailWorkerTimeout.Value < ValidMailMinWorkerTimeout || ref.MailWorkerTimeout.Value > ValidMailMaxWorkerTimeout {
		return ErrMailInvalidWorkerTimeout
	}

	return nil
}
