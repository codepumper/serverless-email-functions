package main

import (
	"context"
	"os"

	// "errors"
	"testing"

	"github.com/resend/resend-go/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockClient struct {
	mock.Mock
}

func (m *MockClient) Emails() *resend.EmailsSvc {
	args := m.Called()
	return args.Get(0).(*resend.EmailsSvc)
}

type MockEmailsService struct {
	mock.Mock
}

func (m *MockEmailsService) Send(params *resend.SendEmailRequest) (*resend.SendEmailResponse, error) {
	args := m.Called(params)
	return args.Get(0).(*resend.SendEmailResponse), args.Error(1)
}

func TestMain(t *testing.T) {

	err := os.Setenv("RESEND_API_KEY", "mock_value")
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	err = os.Setenv("EMAIL_ADDRESS", "mock_value")
	if err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	// Mock the resend client
	mockClient := new(MockClient)
	mockEmailsService := new(MockEmailsService)
	mockClient.On("Emails").Return(mockEmailsService)
	mockEmailsService.On("Send", mock.Anything).Return(&resend.SendEmailResponse{Id: "123"}, nil)

	// Call the function
	resp, err := Main(context.Background(), Event{Name: "Test"})

	// Assert the results
	assert.NoError(t, err)
	assert.Equal(t, "Hello from your email function!", resp.Body)
}

// func TestSendEmailWithRetry(t *testing.T) {
//     // Mock the resend client
//     mockClient := new(MockClient)
//     mockEmailsService := new(MockEmailsService)
//     mockClient.On("Emails").Return(mockEmailsService)

//     // Test successful send
//     mockEmailsService.On("Send", mock.Anything).Return(&resend.SendEmailResponse{Id: "123"}, nil)
//     sent, err := sendEmailWithRetry(mockClient, &resend.SendEmailRequest{}, 5)
//     assert.NoError(t, err)
//     assert.Equal(t, "123", sent.Id)

//     // Test failed send
//     mockEmailsService.On("Send", mock.Anything).Return(nil, errors.New("failed to send email"))
//     sent, err = sendEmailWithRetry(mockClient, &resend.SendEmailRequest{}, 5)
//     assert.Error(t, err)
//     assert.Nil(t, sent)
// }
