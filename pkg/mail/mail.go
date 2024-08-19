package mail

import (
	"crypto/tls"

	"gitlab.devops.telekom.de/schiff/engine/go-breakglass.git/pkg/config"
	"gopkg.in/gomail.v2"
)

type Sender interface {
	Send(receivers []string, subject, body string) error
}

type sender struct {
	dialer *gomail.Dialer
}

func NewSender(cfg config.Config) Sender {
	d := gomail.NewDialer(cfg.Mail.Host, cfg.Mail.Port, cfg.Mail.User, cfg.Mail.Password)
	if cfg.Mail.InsecureSkipVerify {
		d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return &sender{
		dialer: d,
	}
}

func (s *sender) Send(receivers []string, subject, body string) error {
	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", "noreply@schiff.telekom.de", "Das SCHIFF Breakglass")
	msg.SetHeader("Bcc", receivers...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)
	return s.dialer.DialAndSend(msg)
}
