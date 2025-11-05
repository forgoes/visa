package runtime

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"text/template"
)

type Email struct {
	Host     string
	Port     int
	From     string
	Password string
	Server   string
	Template *template.Template
}

func newEmail(config *Config) (*Email, error) {
	server := fmt.Sprintf("%s:%d", config.Deps.Email.Host, config.Deps.Email.Port)

	tmpl, err := template.New("captcha.template").ParseFiles(config.Deps.Email.Template)
	if err != nil {
		return nil, err
	}

	return &Email{
		Host:     config.Deps.Email.Host,
		Port:     config.Deps.Email.Port,
		From:     config.Deps.Email.From,
		Password: config.Deps.Email.Password,
		Server:   server,
		Template: tmpl,
	}, nil
}

func (e *Email) dial() (*smtp.Client, error) {
	conn, err := net.Dial("tcp", e.Server)
	if err != nil {
		return nil, err
	}

	return smtp.NewClient(conn, e.Host)
}

func (e *Email) sendMailTLS(to []string, msg []byte) (err error) {
	c, err := e.dial()
	if err != nil {
		return err
	}
	defer func() {
		_ = c.Close()
	}()

	if err = c.StartTLS(&tls.Config{
		ServerName: e.Host,
	}); err != nil {
		return err
	}

	err = smtp.SendMail(
		e.Server,
		smtp.PlainAuth("", e.From, e.Password, e.Host),
		e.From,
		to,
		msg,
	)
	// https://www.google.com/settings/security/lesssecureapps
	if err != nil {
		return err
	}

	return c.Quit()
}

func (e *Email) Send(captcha string, to string) error {
	data := struct {
		From    string
		To      string
		Subject string
		Captcha string
	}{
		From:    e.From,
		To:      to,
		Subject: "Email Verification",
		Captcha: captcha,
	}

	var buf bytes.Buffer
	err := e.Template.Execute(&buf, data)
	if err != nil {
		return err
	}

	err = e.sendMailTLS(
		[]string{to},
		buf.Bytes(),
	)
	if err != nil {
		return err
	}

	return nil
}
