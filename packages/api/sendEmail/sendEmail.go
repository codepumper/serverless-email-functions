package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"math"
	"math/rand"
	"net/mail"
	"os"
	"strings"
	"time"

	"github.com/resend/resend-go/v2"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Message   string `json:"message"`
}

type Response struct {
	StatusCode int    `json:"statusCode"`
	Body       string `json:"body"`
}

func (e *Event) Validate() error {
	if strings.TrimSpace(e.Email) == "" {
		return errors.New("email cannot be empty")
	}

	_, err := mail.ParseAddress(e.Email)
	if err != nil {
		return errors.New("email is not valid")
	}

	return nil
}

func Main(ctx context.Context, event Event) *Response {
	err := event.Validate()
	if err != nil {
		log.WithError(err).Error("Invalid event data")
		return &Response{
			StatusCode: 400,
			Body:       fmt.Sprintf("%v", err),
		}
	}

	apiKey := os.Getenv("RESEND_API")
	if apiKey == "" {
		log.Error("RESEND_API_KEY environment variable is not set")
		return &Response{
			StatusCode: 500,
			Body:       "RESEND_API_KEY environment variable is not set",
		}
	}

	emailAddress := os.Getenv("TO_EMAIL_ADDRESS")
	if emailAddress == "" {
		log.Error("TO_EMAIL_ADDRESS environment variable is not set")
		return &Response{
			StatusCode: 500,
			Body:       "TO_EMAIL_ADDRESS environment variable is not set",
		}
	}

	client := resend.NewClient(apiKey)

	tmpl := template.Must(template.New("email").Parse(`
        <strong>Name:</strong> {{.FirstName}} {{.LastName}}<br>
        <strong>Email:</strong> {{.Email}}<br>
        <strong>Message:</strong> {{.Message}}
    `))
	var htmlContent strings.Builder
	if err := tmpl.Execute(&htmlContent, event); err != nil {
		log.WithError(err).Error("Failed to construct HTML content")
		return &Response{
			StatusCode: 500,
			Body:       "Failed to construct HTML content",
		}
	}

	params := &resend.SendEmailRequest{
		From:    "Acme <onboarding@resend.dev>",
		To:      []string{emailAddress},
		Html:    htmlContent.String(),
		Subject: "Contact Form Submission",
	}

	maxRetries := 5
	_, err = sendEmailWithRetry(client, params, maxRetries)
	if err != nil {
		log.WithError(err).Error("Failed to send email")
		return &Response{
			StatusCode: 500,
			Body:       "Failed to send email",
		}
	}

	return &Response{
		StatusCode: 200,
		Body:       "Email sent successfully",
	}
}

func sendEmailWithRetry(client *resend.Client, params *resend.SendEmailRequest, maxRetries int) (*resend.SendEmailResponse, error) {
	var sent *resend.SendEmailResponse
	var err error

	for i := 0; i < maxRetries; i++ {
		sent, err = client.Emails.Send(params)
		if err == nil {
			return sent, nil
		}

		log.WithFields(log.Fields{
			"attempt": i + 1,
			"error":   err,
		}).Error("Failed to send email")

		delay := time.Duration(math.Pow(2, float64(i))+float64(rand.Intn(1000))) * time.Millisecond
		log.WithField("delay", delay).Info("Retrying send email")
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("failed to send email after %d attempts: %w", maxRetries, err)
}
