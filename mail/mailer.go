package mail

import (
	"log"
	"net/smtp"
)

func SendMail(recipient string, subject string, messageBody string) error {
	log.Printf("SendMail with \nreciepient: %v\nsubject: %v\nbody: %v\n", recipient, subject, messageBody)
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
	log.Printf("Calling smtp.Sendmail...\n")

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return err
	}

	return nil
}
