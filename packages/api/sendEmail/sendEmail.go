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
	Name string `json:"name"`
	// Method string `json:"method"`
}

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

func Main(ctx context.Context, event Event) (*Response, error) {
	// if event.Method != "POST" {
	// 	return nil, fmt.Errorf("Invalid request method")
	// }

	apiKey := os.Getenv("RESEND_API_KEY")
	if apiKey == "" {
		fmt.Println("RESEND_API_KEY environment variable is not set")
		return nil, fmt.Errorf("RESEND_API_KEY environment variable is not set")
	}

	emailAddress := os.Getenv("EMAIL_ADDRESS")
	if emailAddress == "" {
		fmt.Println("EMAIL_ADDRESS environment variable is not set")
		return nil, fmt.Errorf("EMAIL_ADDRESS environment variable is not set")
	}

	client := resend.NewClient(apiKey)

	params := &resend.SendEmailRequest{
		From:    "Acme <onboarding@resend.dev>",
		To:      []string{emailAddress},
		Html:    "<strong>hello world</strong>",
		Subject: "Hello from Golang",
	}

	maxRetries := 5
	sent, err := sendEmailWithRetry(client, params, maxRetries)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	fmt.Println("Email sent successfully, ID:", sent.Id)

	return &Response{
		Body: fmt.Sprintf("Hello from your email function!"),
	}, nil
}

func sendEmailWithRetry(client *resend.Client, params *resend.SendEmailRequest, maxRetries int) (*resend.SendEmailResponse, error) {
	var sent *resend.SendEmailResponse
	var err error

	for i := 0; i < maxRetries; i++ {
		sent, err = client.Emails.Send(params)
		if err == nil {
			return sent, nil
		}

		// Print the error
		fmt.Println("Attempt", i+1, "failed:", err.Error())

		// Exponential backoff delay
		delay := time.Duration(math.Pow(2, float64(i))) * time.Second
		fmt.Printf("Retrying in %v seconds...\n", delay.Seconds())
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("failed to send email after %d attempts: %w", maxRetries, err)
}
