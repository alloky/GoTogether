package service

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
)

type EmailSender interface {
	SendLinkCode(toEmail, code string) error
}

// SMTPEmailSender sends emails via SMTP.
// Port 465 → implicit TLS (SMTPS).
// Port 587/25 → plain connection with optional STARTTLS + PlainAuth.
// Empty user/password → no authentication (local mail catchers like Mailpit).
type SMTPEmailSender struct {
	host     string
	port     string
	user     string
	password string
	from     string
}

func NewSMTPEmailSender(host, port, user, password, from string) *SMTPEmailSender {
	return &SMTPEmailSender{host: host, port: port, user: user, password: password, from: from}
}

func (s *SMTPEmailSender) SendLinkCode(toEmail, code string) error {
	subject := "GoTogether — Account Linking Code"
	body := fmt.Sprintf(
		"Your one-time code to link your Telegram account:\n\n    %s\n\nThis code expires in 10 minutes.\n\nIf you did not request this, ignore this email.",
		code,
	)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", s.from, toEmail, subject, body)
	addr := net.JoinHostPort(s.host, s.port)

	if s.port == "465" {
		return s.sendTLS(addr, toEmail, msg)
	}
	return s.sendSTARTTLS(addr, toEmail, msg)
}

// sendTLS connects with implicit TLS (port 465 / SMTPS).
func (s *SMTPEmailSender) sendTLS(addr, toEmail, msg string) error {
	tlsCfg := &tls.Config{ServerName: s.host}
	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}

	c, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if s.user != "" && s.password != "" {
		if err := c.Auth(smtp.PlainAuth("", s.user, s.password, s.host)); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	return sendMessage(c, s.from, toEmail, msg)
}

// sendSTARTTLS connects plainly then upgrades with STARTTLS if the server supports it (port 587/25).
func (s *SMTPEmailSender) sendSTARTTLS(addr, toEmail, msg string) error {
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer c.Close()

	// Upgrade to TLS if the server advertises STARTTLS.
	if ok, _ := c.Extension("STARTTLS"); ok {
		if err := c.StartTLS(&tls.Config{ServerName: s.host}); err != nil {
			return fmt.Errorf("starttls: %w", err)
		}
	}

	if s.user != "" && s.password != "" {
		if err := c.Auth(smtp.PlainAuth("", s.user, s.password, s.host)); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	return sendMessage(c, s.from, toEmail, msg)
}

func sendMessage(c *smtp.Client, from, to, msg string) error {
	if err := c.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM: %w", err)
	}
	if err := c.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO: %w", err)
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA: %w", err)
	}
	if _, err := fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("writing message: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("closing data writer: %w", err)
	}
	return c.Quit()
}

// LogEmailSender prints the code to stdout (dev mode when SMTP is not configured).
type LogEmailSender struct{}

func (l *LogEmailSender) SendLinkCode(toEmail, code string) error {
	log.Printf("[DEV] Link code for %s: %s", toEmail, code)
	return nil
}
