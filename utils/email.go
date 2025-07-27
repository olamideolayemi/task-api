package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to string, subject string, body string) error {
	from := os.Getenv("EMAIL_FROM")
	password := os.Getenv("EMAIL_PASSWORD")

	auth := smtp.PlainAuth("", from, password, os.Getenv("SMTP_HOST"))

	msg := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, body))

	err := smtp.SendMail(
		os.Getenv("SMTP_HOST")+":"+os.Getenv("SMTP_PORT"),
		auth,
		from,
		[]string{to},
		msg,
	)

	return err
}
