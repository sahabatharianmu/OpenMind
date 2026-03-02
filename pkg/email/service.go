package email

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"net/smtp"
	"strings"
	texttemplate "text/template"

	"github.com/sahabatharianmu/OpenMind/config"
	"github.com/sahabatharianmu/OpenMind/pkg/logger"
	"go.uber.org/zap"
)

//go:embed templates/*.html templates/*.txt
var templateFS embed.FS

// Parsed template sets (html + text) loaded once at init
var (
	htmlTemplates *htmltemplate.Template
	textTemplates *texttemplate.Template
)

func init() {
	var err error
	htmlTemplates, err = htmltemplate.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		panic("failed to parse HTML email templates: " + err.Error())
	}
	textTemplates, err = texttemplate.ParseFS(templateFS, "templates/*.txt")
	if err != nil {
		panic("failed to parse text email templates: " + err.Error())
	}
}

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

// ---------------------------------------------------------------------------
// Core sending infrastructure
// ---------------------------------------------------------------------------

// SendEmail sends an email using the configured provider
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

// sendViaSMTP sends email via SMTP
func (s *EmailService) sendViaSMTP(to, subject, body string) error {
	smtpConfig := s.config.Email.SMTP
	if smtpConfig.Host == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	port := smtpConfig.Port
	if port == 0 {
		if smtpConfig.TLS {
			port = 587
		} else {
			port = 25 //nolint:mnd
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

	msg := bytes.Buffer{}
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)

	err := smtp.SendMail(addr, auth, from, []string{to}, msg.Bytes())
	if err != nil {
		s.log.Error("Failed to send email via SMTP", zap.Error(err),
			zap.String("to", to), zap.String("subject", subject))
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.log.Info("Email sent successfully", zap.String("to", to), zap.String("subject", subject))
	return nil
}

// sendViaSendGrid sends email via SendGrid (placeholder)
func (s *EmailService) sendViaSendGrid(to, subject, body string) error {
	s.log.Warn("SendGrid integration not implemented yet", zap.String("to", to))
	_ = subject
	_ = body
	return fmt.Errorf("SendGrid integration not implemented")
}

// sendViaAWSES sends email via AWS SES (placeholder)
func (s *EmailService) sendViaAWSES(to, subject, body string) error {
	s.log.Warn("AWS SES integration not implemented yet", zap.String("to", to))
	_ = subject
	_ = body
	return fmt.Errorf("AWS SES integration not implemented")
}

// ---------------------------------------------------------------------------
// Template rendering helpers
// ---------------------------------------------------------------------------

// renderHTML renders an HTML template by name with the given data.
func renderHTML(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := htmlTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("failed to render HTML template %s: %w", name, err)
	}
	return buf.String(), nil
}

// renderText renders a text template by name with the given data.
func renderText(name string, data interface{}) (string, error) {
	var buf bytes.Buffer
	if err := textTemplates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("failed to render text template %s: %w", name, err)
	}
	return buf.String(), nil
}

// sendTemplatedEmail renders and sends an email, falling back to text if HTML fails.
func (s *EmailService) sendTemplatedEmail(to, subject, htmlTmpl, textTmpl string, data interface{}) error {
	htmlBody, err := renderHTML(htmlTmpl, data)
	if err != nil {
		s.log.Error("Failed to render HTML email template", zap.Error(err), zap.String("template", htmlTmpl))
		// Fall through to text
	} else {
		if sendErr := s.SendEmail(to, subject, htmlBody); sendErr == nil {
			return nil
		}
		s.log.Warn("Failed to send HTML email, trying plain text", zap.String("to", to))
	}

	// Fallback to text
	textBody, err := renderText(textTmpl, data)
	if err != nil {
		s.log.Error("Failed to render text email template", zap.Error(err), zap.String("template", textTmpl))
		return fmt.Errorf("failed to render email templates: %w", err)
	}
	return s.SendEmail(to, subject, textBody)
}

// ---------------------------------------------------------------------------
// Public email methods
// ---------------------------------------------------------------------------

// SendWelcomeEmail sends a welcome email to new users after registration
func (s *EmailService) SendWelcomeEmail(to, userName, organizationName, dashboardURL string) error {
	data := map[string]string{
		"UserName":         userName,
		"OrganizationName": organizationName,
		"DashboardURL":     dashboardURL,
	}
	return s.sendTemplatedEmail(to, "Welcome to OpenMind Practice!", "welcome.html", "welcome.txt", data)
}

// SendInvitationEmail sends a team invitation email
func (s *EmailService) SendInvitationEmail(to, inviterName, organizationName, invitationToken, baseURL string) error {
	invitationURL := fmt.Sprintf("%s/accept-invitation?token=%s", baseURL, invitationToken)
	data := map[string]string{
		"InviterName":      inviterName,
		"OrganizationName": organizationName,
		"InvitationURL":    invitationURL,
	}
	subject := fmt.Sprintf("You've been invited to join %s on OpenMind", organizationName)
	return s.sendTemplatedEmail(to, subject, "invitation.html", "invitation.txt", data)
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(to, userName, resetURL string) error {
	data := map[string]string{
		"UserName": userName,
		"ResetURL": resetURL,
	}
	return s.sendTemplatedEmail(to, "Reset Your OpenMind Password", "password_reset.html", "password_reset.txt", data)
}

// SendHandoffRequestEmail sends a patient handoff request notification
func (s *EmailService) SendHandoffRequestEmail(to, patientName, requestingClinicianName, handoffURL string) error {
	data := map[string]string{
		"PatientName":             patientName,
		"RequestingClinicianName": requestingClinicianName,
		"HandoffURL":              handoffURL,
	}
	subject := fmt.Sprintf("Patient Handoff Request: %s", patientName)
	return s.sendTemplatedEmail(to, subject, "handoff_request.html", "handoff_request.txt", data)
}

// SendHandoffApprovedEmail sends a notification when a handoff is approved
func (s *EmailService) SendHandoffApprovedEmail(to, patientName, receivingClinicianName string) error {
	data := map[string]string{
		"PatientName":            patientName,
		"ReceivingClinicianName": receivingClinicianName,
	}
	subject := fmt.Sprintf("Patient Handoff Approved: %s", patientName)

	htmlBody, err := renderHTML("handoff_approved.html", data)
	if err != nil {
		s.log.Error("Failed to render handoff approved HTML", zap.Error(err))
		return s.SendEmail(to, subject, fmt.Sprintf("Patient handoff for %s has been approved by %s.", patientName, receivingClinicianName))
	}
	return s.SendEmail(to, subject, htmlBody)
}

// SendHandoffRejectedEmail sends a notification when a handoff is rejected
func (s *EmailService) SendHandoffRejectedEmail(to, patientName, rejectingClinicianName, reason string) error {
	data := map[string]string{
		"PatientName":            patientName,
		"RejectingClinicianName": rejectingClinicianName,
		"Reason":                 reason,
	}
	subject := fmt.Sprintf("Patient Handoff Rejected: %s", patientName)

	htmlBody, err := renderHTML("handoff_rejected.html", data)
	if err != nil {
		s.log.Error("Failed to render handoff rejected HTML", zap.Error(err))
		return s.SendEmail(to, subject, fmt.Sprintf("Patient handoff for %s has been rejected by %s.", patientName, rejectingClinicianName))
	}
	return s.SendEmail(to, subject, htmlBody)
}
