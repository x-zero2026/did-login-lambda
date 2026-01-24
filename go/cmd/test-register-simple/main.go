package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("=== Test Register Handler ===")
	fmt.Printf("Request Body: %s\n", request.Body)
	
	// Parse request body
	var req RegisterRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		body, _ := json.Marshal(map[string]interface{}{
			"success": false,
			"error":   "Invalid request body",
		})
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: string(body),
		}, nil
	}
	
	fmt.Printf("Parsed: email=%s, username=%s\n", req.Email, req.Username)
	
	// Return success
	body, _ := json.Marshal(map[string]interface{}{
		"success":  true,
		"message":  "Test successful",
		"email":    req.Email,
		"username": req.Username,
	})
	
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
