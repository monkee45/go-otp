package mail

import (
	"log/slog"
	"net/smtp"
)

// *** SendMail ***
// connects to Google Mail API. Uses the passed parameters to construct an email
// It call stmp.SendMail to send the email and returns any error

func SendMail(recipient string, subject string, messageBody string, logger *slog.Logger) {
	// SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Sender credentials
	from := "michael.tom.walsh@gmail.com"
	password := "uwgm hegg megl ikov"
	username := "michael.tom.walsh@gmail.com"

	// Recipient
	to := []string{recipient}

	// Message : RFC 822 standard
	message := []byte("Subject: " + subject + "\r\n" +
		"\r\n" + messageBody + "\r\n")

	// Authentication
	auth := smtp.PlainAuth("", username, password, smtpHost)

	// Send email

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		logger.Error("mail.SendMail()", "Failed to send email", err)
	}

}
