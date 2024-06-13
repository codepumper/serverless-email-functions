package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/resend/resend-go/v2"
	log "github.com/sirupsen/logrus"
)

type Event struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

func (e *Event) Validate() error {
	if e.Email == "" {
		return errors.New("email cannot be empty")
	}
	return nil
}

func Main(ctx context.Context, event Event) (*Response, error) {
	err := event.Validate()
	if err != nil {
		log.WithError(err).Error("Invalid event data")
		return nil, err
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		log.Error("RESEND_API_KEY environment variable is not set")
		return nil, errors.New("RESEND_API_KEY environment variable is not set")
	}

	audienceId := os.Getenv("AUDIENCE_ID")
	if audienceId == "" {
		log.Error("AUDIENCE_ID environment variable is not set")
		return nil, errors.New("AUDIENCE_ID environment variable is not set")
	}

	client := resend.NewClient(apiKey)

	params := &resend.CreateContactRequest{
		Email:        event.Email,
		FirstName:    event.FirstName,
		LastName:     event.LastName,
		Unsubscribed: false,
		AudienceId:   audienceId,
	}

	maxRetries := 5
	added, err := addContactWithRetry(client, params, maxRetries)
	if err != nil {
		log.WithError(err).Error("Failed to add contact")
		return nil, err
	}

	log.WithField("ID", added.Id).Info("Contact added successfully")

	return &Response{
		Body: "Hello from your add contact function!",
	}, nil
}

func addContactWithRetry(client *resend.Client, params *resend.CreateContactRequest, maxRetries int) (*resend.CreateContactResponse, error) {
	var added resend.CreateContactResponse
	var err error

	for i := 0; i < maxRetries; i++ {
		added, err = client.Contacts.Create(params)
		if err == nil {
			return &added, nil
		}

		log.WithFields(log.Fields{
			"attempt": i + 1,
			"error":   err,
		}).Error("Failed to add contact")

		delay := time.Duration(math.Pow(2, float64(i))+float64(rand.Intn(1000))) * time.Millisecond
		log.WithField("delay", delay).Info("Retrying add contact")
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("failed to add contact after %d attempts: %w", maxRetries, err)
}
