package mail

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
