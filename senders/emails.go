package senders

import (
	"bytes"
	"html/template"
	"log"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"gopkg.in/gomail.v2"
)

func sendEmail(recipientEmail, subject, templateFile string, data interface{}) error {
	cfg := config.GetConfig()

	// Parse the template file
	tmpl, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	// Execute the template with the context and set it as the body of the email
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		log.Fatal("Error executing template:", err)
	}

	// Create a new email message
	m := gomail.NewMessage()
	m.SetHeader("From", "TechnoTrades <"+cfg.MailSenderEmail+">")
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body.String())

	// Create a new SMTP client
	d := gomail.NewDialer(cfg.MailSenderHost, cfg.MailSenderPort, cfg.MailSenderEmail, cfg.MailSenderPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		log.Fatal("Error sending email:", err)
	}

	return nil
}

// Struct for verification email data
type VerificationEmailData struct {
	VerificationCode int64
}

// Function to send verification email
func SendVerificationEmail(recipientEmail string, verificationCode int64) error {
	// Prepare the data
	data := VerificationEmailData{
		VerificationCode: verificationCode,
	}

	// Template file and subject for verification email
	templateFile := "utils/mail/templates/account-verification.html"
	subject := "Please Verify your Account"

	// Call the generic sendEmail function
	return sendEmail(recipientEmail, subject, templateFile, data)
}

// Struct for reset password email data
type ResetPasswordEmailData struct {
	Name  string
	Token string
}

// Function to send reset password email
func SendResetPasswordEmail(recipientEmail, name, token string) error {
	// Prepare the data
	data := ResetPasswordEmailData{
		Name:  name,
		Token: token,
	}

	// Template file and subject for reset password email
	templateFile := "utils/mail/templates/reset-password.html"
	subject := "Reset Your Password"

	// Call the generic SendEmail function
	return sendEmail(recipientEmail, subject, templateFile, data)
}

// Struct for reset password email data
type UpdateEmailEmailData struct {
	Name  string
	Token string
}

// Function to send reset password email
func SendUpdateEmailEmail(recipientEmail, name, token string) error {
	// Prepare the data
	data := UpdateEmailEmailData{
		Name:  name,
		Token: token,
	}

	// Template file and subject for reset password email
	templateFile := "utils/mail/templates/update-email.html"
	subject := "Update Your Email"

	// Call the generic SendEmail function
	return sendEmail(recipientEmail, subject, templateFile, data)
}
