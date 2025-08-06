package config

import (
	"errors"
	"os"
	"testing"
	"time"
)

func TestNewMailConfig(t *testing.T) {
	config := NewMailConfig()

	if config.SMTPHost.Value != DefaultMailSMTPHost {
		t.Errorf("Expected SMTPHost to be %s, got %s", DefaultMailSMTPHost, config.SMTPHost.Value)
	}
	if config.SMTPPort.Value != DefaultMailSMTPPort {
		t.Errorf("Expected SMTPPort to be %d, got %d", DefaultMailSMTPPort, config.SMTPPort.Value)
	}
	if config.SMTPUsername.Value != DefaultMailSMTPUsername {
		t.Errorf("Expected SMTPUsername to be %s, got %s", DefaultMailSMTPUsername, config.SMTPUsername.Value)
	}
	if config.SMTPPassword.Value != DefaultMailSMTPPassword {
		t.Errorf("Expected SMTPPassword to be %s, got %s", DefaultMailSMTPPassword, config.SMTPPassword.Value)
	}
	if config.SenderName.Value != DefaultMailSenderName {
		t.Errorf("Expected SenderName to be %s, got %s", DefaultMailSenderName, config.SenderName.Value)
	}
	if config.SenderAddress.Value != DefaultMailSenderAddress {
		t.Errorf("Expected SenderAddress to be %s, got %s", DefaultMailSenderAddress, config.SenderAddress.Value)
	}
	if config.APIURL.Value != DefaultMailAPIEndpoint {
		t.Errorf("Expected APIURL to be %s, got %s", DefaultMailAPIEndpoint, config.APIURL.Value)
	}
	if config.APIKey.Value != DefaultMailAPIKey {
		t.Errorf("Expected APIKey to be %s, got %s", DefaultMailAPIKey, config.APIKey.Value)
	}
	if config.MailSender.Value != DefaultMailSender {
		t.Errorf("Expected MailSender to be %s, got %s", DefaultMailSender, config.MailSender.Value)
	}
	if config.MailWorkerCount.Value != DefaultMailWorkerCount {
		t.Errorf("Expected MailWorkerCount to be %d, got %d", DefaultMailWorkerCount, config.MailWorkerCount.Value)
	}
	if config.MailWorkerTimeout.Value != DefaultMailWorkerTimeout {
		t.Errorf("Expected MailWorkerTimeout to be %v, got %v", DefaultMailWorkerTimeout, config.MailWorkerTimeout.Value)
	}
}

func TestParseEnvVars_mail(t *testing.T) {
	os.Setenv("MAIL_SMTP_HOST", "smtp.example.com")
	os.Setenv("MAIL_SMTP_PORT", "465")
	os.Setenv("MAIL_SMTP_USERNAME", "testuser")
	os.Setenv("MAIL_SMTP_PASSWORD", "testpass")
	os.Setenv("MAIL_SENDER_NAME", "Test Sender")
	os.Setenv("MAIL_SENDER_ADDRESS", "test@example.com")
	os.Setenv("MAIL_API_URL", "https://api.mailgun.net")
	os.Setenv("MAIL_API_KEY", "test_api_key")
	os.Setenv("MAIL_SENDER", "mailgun")
	os.Setenv("MAIL_WORKER_COUNT", "10")
	os.Setenv("MAIL_WORKER_TIMEOUT", "8s")

	config := NewMailConfig()
	config.ParseEnvVars()

	if config.SMTPHost.Value != "smtp.example.com" {
		t.Errorf("Expected SMTPHost to be smtp.example.com, got %s", config.SMTPHost.Value)
	}
	if config.SMTPPort.Value != 465 {
		t.Errorf("Expected SMTPPort to be 465, got %d", config.SMTPPort.Value)
	}
	if config.SMTPUsername.Value != "testuser" {
		t.Errorf("Expected SMTPUsername to be testuser, got %s", config.SMTPUsername.Value)
	}
	if config.SMTPPassword.Value != "testpass" {
		t.Errorf("Expected SMTPPassword to be testpass, got %s", config.SMTPPassword.Value)
	}
	if config.SenderName.Value != "Test Sender" {
		t.Errorf("Expected SenderName to be Test Sender, got %s", config.SenderName.Value)
	}
	if config.SenderAddress.Value != "test@example.com" {
		t.Errorf("Expected SenderAddress to be test@example.com, got %s", config.SenderAddress.Value)
	}
	if config.APIURL.Value != "https://api.mailgun.net" {
		t.Errorf("Expected APIURL to be https://api.mailgun.net, got %s", config.APIURL.Value)
	}
	if config.APIKey.Value != "test_api_key" {
		t.Errorf("Expected APIKey to be test_api_key, got %s", config.APIKey.Value)
	}
	if config.MailSender.Value != "mailgun" {
		t.Errorf("Expected MailSender to be mailgun, got %s", config.MailSender.Value)
	}
	if config.MailWorkerCount.Value != 10 {
		t.Errorf("Expected MailWorkerCount to be 10, got %d", config.MailWorkerCount.Value)
	}
	if config.MailWorkerTimeout.Value != 8*time.Second {
		t.Errorf("Expected MailWorkerTimeout to be 8s, got %v", config.MailWorkerTimeout.Value)
	}

	// Clean up environment variables
	os.Unsetenv("MAIL_SMTP_HOST")
	os.Unsetenv("MAIL_SMTP_PORT")
	os.Unsetenv("MAIL_SMTP_USERNAME")
	os.Unsetenv("MAIL_SMTP_PASSWORD")
	os.Unsetenv("MAIL_SENDER_NAME")
	os.Unsetenv("MAIL_SENDER_ADDRESS")
	os.Unsetenv("MAIL_API_URL")
	os.Unsetenv("MAIL_API_KEY")
	os.Unsetenv("MAIL_SENDER")
	os.Unsetenv("MAIL_WORKER_COUNT")
	os.Unsetenv("MAIL_WORKER_TIMEOUT")
}

func TestValidate_mail(t *testing.T) {
	config := NewMailConfig()

	// Test valid SMTP configuration
	config.SMTPHost.Value = "smtp.example.com"
	config.SMTPUsername.Value = "testuser"
	config.SMTPPassword.Value = "testpass"
	config.MailSender.Value = "smtp"
	config.APIURL.Value = "https://api.mailgun.net" // Set to pass the first validation

	err := config.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid SMTP config, got %v", err)
	}

	// Test invalid SMTP port
	config.SMTPPort.Value = 999
	err = config.Validate()
	var invalidErr *InvalidConfigurationError
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.smtp.port" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.smtp.port', got %v", err)
	}
	config.SMTPPort.Value = DefaultMailSMTPPort

	// Test invalid mail sender
	config.MailSender.Value = "invalid"
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.sender" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.sender', got %v", err)
	}
	config.MailSender.Value = "smtp"

	// Test missing SMTP username for smtp sender
	config.SMTPUsername.Value = ""
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.smtp.username" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.smtp.username', got %v", err)
	}
	config.SMTPUsername.Value = "testuser"

	// Test missing SMTP password when username is set
	config.SMTPPassword.Value = ""
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.smtp.password" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.smtp.password', got %v", err)
	}
	config.SMTPPassword.Value = "testpass"

	// Test valid mailgun configuration
	config.MailSender.Value = "mailgun"
	config.APIURL.Value = "https://api.mailgun.net"
	config.APIKey.Value = "test_api_key"
	err = config.Validate()
	if err != nil {
		t.Errorf("Expected no error for valid mailgun config, got %v", err)
	}

	// Test invalid API URL for mailgun
	config.APIURL.Value = ":/invalid-url"
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.api.url" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.api.url', got %v", err)
	}
	config.APIURL.Value = "https://api.mailgun.net"

	// Test missing API key for mailgun
	config.APIKey.Value = ""
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.api.key" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.api.key', got %v", err)
	}
	config.APIKey.Value = "test_api_key"

	// Test empty sender name
	config.SenderName.Value = ""
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.sender.name" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.sender.name', got %v", err)
	}
	config.SenderName.Value = DefaultMailSenderName

	// Test invalid sender address
	config.SenderAddress.Value = "invalid-email"
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.sender.address" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.sender.address', got %v", err)
	}
	config.SenderAddress.Value = DefaultMailSenderAddress

	// Test invalid worker count (too low)
	config.MailWorkerCount.Value = 0
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.worker.count" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.worker.count', got %v", err)
	}
	config.MailWorkerCount.Value = DefaultMailWorkerCount

	// Test invalid worker timeout (too short)
	config.MailWorkerTimeout.Value = 500 * time.Millisecond
	err = config.Validate()
	if err == nil || !errors.As(err, &invalidErr) || invalidErr.Field != "mail.worker.timeout" {
		t.Errorf("Expected InvalidConfigurationError with field 'mail.worker.timeout', got %v", err)
	}
	config.MailWorkerTimeout.Value = DefaultMailWorkerTimeout
}
