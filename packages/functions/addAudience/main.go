package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/resend/resend-go/v2"
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

func Main(ctx context.Context, event Event) (*Response, error) {
	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		fmt.Println("RESEND_API_KEY environment variable is not set")
		return nil, fmt.Errorf("RESEND_API_KEY environment variable is not set")
	}

	audienceId := os.Getenv("AUDIENCE_ID")
	if audienceId == "" {
		fmt.Println("AUDIENCE_ID environment variable is not set")
		return nil, fmt.Errorf("AUDIENCE_ID environment variable is not set")
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
		fmt.Println(err.Error())
		return nil, err
	}

	fmt.Println("Contact added successfully, ID:", added.Id)

	return &Response{
		Body: fmt.Sprintf("Hello from your add contact function!"),
	}, nil
}

func addContactWithRetry(client *resend.Client, params *resend.CreateContactRequest, maxRetries int) (*resend.CreateContactResponse, error) {
	var err error

	for i := 0; i < maxRetries; i++ {
		contact, err := client.Contacts.Create(params)
		if err == nil {
			return &contact, nil
		}

		// Print the error
		fmt.Println("Attempt", i+1, "failed:", err.Error())

		// Exponential backoff delay
		delay := time.Duration(math.Pow(2, float64(i))) * time.Second
		fmt.Printf("Retrying in %v seconds...\n", delay.Seconds())
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("failed to add contact after %d attempts: %w", maxRetries, err)
}
