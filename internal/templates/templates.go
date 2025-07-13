// Package templates provides functionality to render email templates for account verification.
package templates

import (
	"bytes"
	"errors"
	"html/template"
)

var (
	ErrInvalidAccountVerificationConf = errors.New("invalid account verification conf")
	ErrInvalidVerificationAPIEndpoint = errors.New("invalid verification API endpoint")
	ErrInvalidVerificationToken       = errors.New("invalid verification token")
	ErrInvalidVerificationTTL         = errors.New("invalid verification TTL")
	ErrInvalidUserName                = errors.New("invalid user name")
)

const accountVerificationHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Account Verification</title>
</head>

<body>
    <h1>Account Verification</h1>
    <p>Dear User {{.UserName}},</p>
    <p>Thank you for signing up!</p>
    <p>To complete your registration, please verify your account.</p>
    <p>We are excited to have you on board!</p>
    <p>To verify your account, please click the link below:</p>
    <a href="{{.VerificationAPIEndpoint}}/{{.VerificationToken}}">Verify Account</a>
    <br/>
    <h3>Token will expire in {{.VerificationTTL}}</h3>
</body>
</html>
`

const accountVerificationText = `
Account Verification

Dear User {{.UserName}},
Thank you for signing up!
To complete your registration, please verify your account.
We are excited to have you on board!
To verify your account, please click the link below:

{{.VerificationAPIEndpoint}}/{{.VerificationToken}}

Token will expire in {{.VerificationTTL}}
`

type EmailAccountVerificationConf struct {
	VerificationAPIEndpoint string
	VerificationToken       string
	VerificationTTL         string
	UserName                string
	HTML                    bool
}
type EmailAccountVerification struct {
	verificationAPIEndpoint string
	verificationToken       string
	verificationTTL         string
	userName                string
	html                    bool
}

func NewEmailAccountVerification(conf *EmailAccountVerificationConf) (*EmailAccountVerification, error) {
	if conf == nil {
		return nil, ErrInvalidAccountVerificationConf
	}

	if conf.VerificationAPIEndpoint == "" {
		return nil, ErrInvalidVerificationAPIEndpoint
	}

	if conf.VerificationToken == "" {
		return nil, ErrInvalidVerificationToken
	}

	if conf.VerificationTTL == "" {
		return nil, ErrInvalidVerificationTTL
	}

	if conf.UserName == "" {
		return nil, ErrInvalidUserName
	}

	return &EmailAccountVerification{
		verificationAPIEndpoint: conf.VerificationAPIEndpoint,
		verificationToken:       conf.VerificationToken,
		verificationTTL:         conf.VerificationTTL,
		userName:                conf.UserName,
		html:                    conf.HTML,
	}, nil
}

func (e *EmailAccountVerification) Render() string {
	var tmplType string

	if e.html {
		tmplType = accountVerificationHTML
	} else {
		tmplType = accountVerificationText
	}

	data := struct {
		VerificationAPIEndpoint string
		VerificationToken       string
		VerificationTTL         string
		UserName                string
	}{
		VerificationAPIEndpoint: e.verificationAPIEndpoint,
		VerificationToken:       e.verificationToken,
		VerificationTTL:         e.verificationTTL,
		UserName:                e.userName,
	}

	tmpl, err := template.New("accountVerification").Parse(tmplType)
	if err != nil {
		panic(err)
	}

	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, data)
	if err != nil {
		panic(err)
	}

	return tpl.String()
}
