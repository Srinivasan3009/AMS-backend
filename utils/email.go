package utils

import (
	"fmt"
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// SendGradeNotification emails a student that their grade for a course has been posted/updated.
func SendGradeNotification(
	toEmail,
	studentName,
	courseNo,
	courseName,
	term,
	grade string,
) error {

	apiKey := os.Getenv("SENDGRID_API_KEY")
	fromEmail := os.Getenv("SMTP_USER")
	fromName := os.Getenv("SMTP_FROM_NAME")

	if fromName == "" {
		fromName = "Anna University Portal"
	}

	if apiKey == "" {
		return fmt.Errorf("SENDGRID_API_KEY is not configured")
	}

	if fromEmail == "" {
		return fmt.Errorf("SMTP_USER is not configured")
	}

	subject := fmt.Sprintf(
		"Grade Posted: %s (%s)",
		courseName,
		courseNo,
	)

	body := fmt.Sprintf(
		"Dear %s,\n\n"+
			"Your grade for the following course has been posted:\n\n"+
			"Course: %s - %s\n"+
			"Term: %s\n"+
			"Grade: %s\n\n"+
			"You can view this in your student portal under Academic Information or Grade Viewer.\n\n"+
			"Regards,\n%s",
		studentName,
		courseNo,
		courseName,
		term,
		grade,
		fromName,
	)

	from := mail.NewEmail(fromName, fromEmail)
	to := mail.NewEmail(studentName, toEmail)

	message := mail.NewSingleEmail(
		from,
		subject,
		to,
		body,
		body,
	)

	client := sendgrid.NewSendClient(apiKey)

	response, err := client.Send(message)

	if err != nil {
		return fmt.Errorf("SendGrid error: %w", err)
	}

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf(
			"SendGrid returned status %d: %s",
			response.StatusCode,
			response.Body,
		)
	}

	return nil
}
