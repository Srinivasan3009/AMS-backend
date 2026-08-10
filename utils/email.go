package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

// SendGradeNotification emails a student that their grade for a course has been posted/updated.
func SendGradeNotification(toEmail, studentName, courseNo, courseName, term, grade string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	password := os.Getenv("SMTP_PASSWORD")
	fromName := os.Getenv("SMTP_FROM_NAME")
	if fromName == "" {
		fromName = "Anna University Portal"
	}

	if host == "" || port == "" || user == "" || password == "" {
		return fmt.Errorf("SMTP is not configured")
	}

	subject := fmt.Sprintf("Grade Posted: %s (%s)", courseName, courseNo)
	body := fmt.Sprintf(
		"Dear %s,\r\n\r\n"+
			"Your grade for the following course has been posted:\r\n\r\n"+
			"Course: %s - %s\r\n"+
			"Term: %s\r\n"+
			"Grade: %s\r\n\r\n"+
			"You can view this in your student portal under Academic Information or Grade Viewer.\r\n\r\n"+
			"Regards,\r\n%s",
		studentName, courseNo, courseName, term, grade, fromName,
	)

	msg := fmt.Sprintf(
		"From: %s <%s>\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"UTF-8\"\r\n\r\n%s",
		fromName, user, toEmail, subject, body,
	)

	auth := smtp.PlainAuth("", user, password, host)
	addr := host + ":" + port

	return smtp.SendMail(addr, auth, user, []string{toEmail}, []byte(msg))
}
