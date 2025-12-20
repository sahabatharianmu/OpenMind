package email

import (
	"bytes"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
)

// EmailService handles email sending operations
type EmailService struct {
	config *config.Config
	log    logger.Logger
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.Config, log logger.Logger) *EmailService {
	return &EmailService{
		config: cfg,
		log:    log,
	}
}

// SendEmail sends an email
func (s *EmailService) SendEmail(to, subject, body string) error {
	switch strings.ToLower(s.config.Email.Provider) {
	case "smtp":
		return s.sendViaSMTP(to, subject, body)
	case "sendgrid":
		return s.sendViaSendGrid(to, subject, body)
	case "aws_ses":
		return s.sendViaAWSES(to, subject, body)
	default:
		s.log.Warn("Email provider not configured, logging email instead", zap.String("provider", s.config.Email.Provider))
		s.log.Info("Email would be sent", zap.String("to", to), zap.String("subject", subject))
		return nil
	}
}

// SendInvitationEmail sends a team invitation email
func (s *EmailService) SendInvitationEmail(to, inviterName, organizationName, invitationToken, baseURL string) error {
	invitationURL := fmt.Sprintf("%s/accept-invitation?token=%s", baseURL, invitationToken)

	subject := fmt.Sprintf("You've been invited to join %s on OpenMind", organizationName)

	// HTML email template
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>Team Invitation</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
	<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
		<h2 style="color: #2c3e50;">You've been invited to join a team!</h2>
		<p>Hello,</p>
		<p><strong>%s</strong> has invited you to join <strong>%s</strong> on OpenMind Practice.</p>
		<p>Click the button below to accept the invitation:</p>
		<div style="text-align: center; margin: 30px 0;">
			<a href="%s" style="background-color: #3498db; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; display: inline-block;">Accept Invitation</a>
		</div>
		<p>Or copy and paste this link into your browser:</p>
		<p style="word-break: break-all; color: #7f8c8d;">%s</p>
		<p style="color: #95a5a6; font-size: 12px; margin-top: 30px;">This invitation will expire in 7 days. If you didn't expect this invitation, you can safely ignore this email.</p>
	</div>
</body>
</html>
`, inviterName, organizationName, invitationURL, invitationURL)

	// Plain text fallback
	textBody := fmt.Sprintf(`
You've been invited to join a team!

%s has invited you to join %s on OpenMind Practice.

Accept the invitation by clicking this link:
%s

This invitation will expire in 7 days. If you didn't expect this invitation, you can safely ignore this email.
`, inviterName, organizationName, invitationURL)

	// Try HTML first, fallback to plain text
	if err := s.SendEmail(to, subject, htmlBody); err != nil {
		s.log.Warn("Failed to send HTML email, trying plain text", zap.Error(err))
		return s.SendEmail(to, subject, textBody)
	}

	return nil
}

// sendViaSMTP sends email via SMTP
func (s *EmailService) sendViaSMTP(to, subject, body string) error {
	smtpConfig := s.config.Email.SMTP
	if smtpConfig.Host == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	// Set default port if not specified
	port := smtpConfig.Port
	if port == 0 {
		if smtpConfig.TLS {
			port = 587
		} else {
			port = 25
		}
	}

	addr := fmt.Sprintf("%s:%d", smtpConfig.Host, port)
	auth := smtp.PlainAuth("", smtpConfig.Username, smtpConfig.Password, smtpConfig.Host)

	from := s.config.Email.FromEmail
	if from == "" {
		from = smtpConfig.Username
	}

	fromName := s.config.Email.FromName
	if fromName == "" {
		fromName = "OpenMind"
	}

	// Build email message
	msg := bytes.Buffer{}
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Send email
	err := smtp.SendMail(addr, auth, from, []string{to}, msg.Bytes())
	if err != nil {
		s.log.Error("Failed to send email via SMTP", zap.Error(err),
			zap.String("to", to), zap.String("subject", subject))
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.log.Info("Email sent successfully", zap.String("to", to), zap.String("subject", subject))
	return nil
}

// sendViaSendGrid sends email via SendGrid (placeholder - needs SendGrid SDK)
func (s *EmailService) sendViaSendGrid(to, subject, body string) error {
	s.log.Warn("SendGrid integration not implemented yet", zap.String("to", to))
	return fmt.Errorf("SendGrid integration not implemented")
}

// sendViaAWSES sends email via AWS SES (placeholder - needs AWS SDK)
func (s *EmailService) sendViaAWSES(to, subject, body string) error {
	s.log.Warn("AWS SES integration not implemented yet", zap.String("to", to))
	return fmt.Errorf("AWS SES integration not implemented")
}
