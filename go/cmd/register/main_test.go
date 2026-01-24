package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandler(t *testing.T) {
	// Set environment variables
	os.Setenv("SUPABASE_URL", "https://rbpsksuuvtzmathnmyxn.supabase.co")
	os.Setenv("DB_PASSWORD", "iPass4xz2026!")
	os.Setenv("JWT_SECRET", "Ia7gdt+1znW6j9I9XXLg+//MbKYIMa3HW5X7Eqd3gho=")
	os.Setenv("JWT_EXPIRY", "168h")

	// Create test request
	reqBody := RegisterRequest{
		Email:    "test@example.com",
		Username: "testuser",
		Password: "password123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	request := events.APIGatewayProxyRequest{
		Body: string(bodyBytes),
	}

	// Call handler
	response, err := handler(context.Background(), request)
	
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	t.Logf("Status Code: %d", response.StatusCode)
	t.Logf("Body: %s", response.Body)
	
	if response.StatusCode != 200 && response.StatusCode != 409 {
		t.Errorf("Expected status 200 or 409, got %d", response.StatusCode)
	}
}
