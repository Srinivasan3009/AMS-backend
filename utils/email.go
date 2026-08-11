package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type resendEmailRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text"`
}

// SendGradeNotification emails a student that their grade for a course has been posted/updated.
// Uses Resend's HTTP API (not SMTP) - works reliably on hosts like Render that block
// outbound SMTP ports, and avoids DMARC alignment issues since we send from Resend's
// own sandbox domain (onboarding@resend.dev) rather than impersonating a Gmail address.
func SendGradeNotification(toEmail, studentName, courseNo, courseName, term, grade string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not configured")
	}

	subject := fmt.Sprintf("Grade Posted: %s (%s)", courseName, courseNo)
	body := fmt.Sprintf(
		"Dear %s,\n\n"+
			"Your grade for the following course has been posted:\n\n"+
			"Course: %s - %s\n"+
			"Term: %s\n"+
			"Grade: %s\n\n"+
			"You can view this in your student portal under Academic Information or Grade Viewer.\n\n"+
			"Regards,\nAnna University Portal",
		studentName, courseNo, courseName, term, grade,
	)

	reqBody := resendEmailRequest{
		From:    "Anna University Portal <onboarding@resend.dev>",
		To:      []string{toEmail},
		Subject: subject,
		Text:    body,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("could not build email request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("could not create email request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send email: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
