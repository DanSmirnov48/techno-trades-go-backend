package senders

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/DanSmirnov48/techno-trades-go-backend/config"
	"github.com/DanSmirnov48/techno-trades-go-backend/models"
	"gopkg.in/gomail.v2"
)

type EmailContext struct {
	Name string
	Otp  *uint32
}

type EmailType string

const (
	EmailWelcome              EmailType = "welcome"
	EmailActivate             EmailType = "activate"
	EmailOtpLogin             EmailType = "otp-login"
	EmailResetPassword        EmailType = "reset-password"
	EmailResetPasswordSuccess EmailType = "reset-password-success"
)

func sortEmail(emailType EmailType, code *uint32) map[string]interface{} {
	data := make(map[string]interface{})

	switch emailType {
	case EmailWelcome:
		data["template_file"] = "senders/templates/welcome.html"
		data["subject"] = "Account verified"

	case EmailActivate:
		data["template_file"] = "senders/templates/account-verification.html"
		data["subject"] = "Activate your account"
		data["otp"] = code

	case EmailOtpLogin:
		data["template_file"] = "senders/templates/otp-login.html"
		data["subject"] = "One time Login Password"
		data["otp"] = code

	case EmailResetPassword:
		data["template_file"] = "senders/templates/reset-password.html"
		data["subject"] = "Reset your password"
		data["otp"] = code

	case EmailResetPasswordSuccess:
		data["template_file"] = "senders/templates/reset-password-success.html"
		data["subject"] = "Password reset successfully"
		data["otp"] = code
	}
	return data
}

func SendEmail(user *models.User, emailType EmailType, code *uint32) {
	if os.Getenv("ENVIRONMENT") == "TESTING" {
		return
	}

	cfg := config.GetConfig()

	emailData := sortEmail(emailType, code)
	templateFile := emailData["template_file"]
	subject := emailData["subject"]

	// Create a context with dynamic data
	data := EmailContext{
		Name: user.FirstName,
	}
	if otp, ok := emailData["otp"]; ok {
		code := otp.(*uint32)
		data.Otp = code
	}

	// Read the HTML file content
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Println("Unable to identify current directory (needed to load templates)", os.Stderr)
		os.Exit(1)
	}
	basepath := filepath.Dir(file)
	tempfile := fmt.Sprintf("../%s", templateFile.(string))
	htmlContent, err := os.ReadFile(filepath.Join(basepath, tempfile))
	if err != nil {
		log.Fatal("Error reading HTML file:", err)
	}

	// Create a new template from the HTML file content
	tmpl, err := template.New("email_template").Parse(string(htmlContent))
	if err != nil {
		log.Fatal("Error parsing template:", err)
	}

	// Execute the template with the context and set it as the body of the email
	var bodyContent bytes.Buffer
	if err := tmpl.Execute(&bodyContent, data); err != nil {
		log.Fatal("Error executing template:", err)
	}

	// Create a new message
	m := gomail.NewMessage()
	m.SetHeader("From", "TechnoTrades <"+cfg.MailSenderEmail+">")
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", subject.(string))
	m.SetBody("text/html", bodyContent.String())

	// Create a new SMTP client
	d := gomail.NewDialer(cfg.MailSenderHost, cfg.MailSenderPort, cfg.MailSenderEmail, cfg.MailSenderPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		log.Fatal("Error sending email:", err)
	}
}
