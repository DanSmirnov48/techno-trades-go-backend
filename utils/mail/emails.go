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
