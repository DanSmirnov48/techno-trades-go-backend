package mail

import (
	"bytes"
	"log"
	"os"

	"strconv"
	"text/template"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

type EmailData struct {
	Name            string
	ConfirmationURL string
}

var host string
var port int
var username string
var password string

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}

	host = os.Getenv("EMAIL_HOST")
	port, _ = strconv.Atoi(os.Getenv("EMAIL_PORT"))
	username = os.Getenv("EMAIL_USERNAME")
	password = os.Getenv("EMAIL_PASSWORD")
}

func SendEmail(recipientEmail string, data EmailData) error {
	// Load the HTML template file
	tmpl, err := template.ParseFiles("utils/mail/templates/email_template.html")
	if err != nil {
		return err
	}

	// Buffer to hold the rendered HTML email body
	var body bytes.Buffer

	// Render the template with dynamic data into the buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return err
	}

	// Create a new email message
	m := gomail.NewMessage()

	// Set the sender, recipient, and subject
	m.SetHeader("From", "TechnoTrades <"+username+">")
	m.SetHeader("To", recipientEmail)
	m.SetHeader("Subject", "Please Confirm Your Email")

	// Set the rendered HTML body
	m.SetBody("text/html", body.String())

	// Set up the SMTP server configuration
	d := gomail.NewDialer(host, port, username, password)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
