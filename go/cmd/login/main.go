package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/db"
	"github.com/x-zero/did-login/pkg/response"
)

type LoginRequest struct {
	Identifier string `json:"identifier"` // email or username
	Password   string `json:"password"`
}

type LoginResponse struct {
	DID      string `json:"did"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize database
	if err := db.InitDB(); err != nil {
		return response.Error(500, "Database connection failed")
	}

	// Parse request body
	var req LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.Error(400, "Invalid request body")
	}

	// Validate input
	if req.Identifier == "" || req.Password == "" {
		return response.Error(400, "Identifier and password are required")
	}

	// Query user by email or username
	pool := db.GetPool()
	var did, username, passwordHash string
	err := pool.QueryRow(ctx,
		"SELECT did, username, password_hash FROM users WHERE email = $1 OR username = $1",
		req.Identifier,
	).Scan(&did, &username, &passwordHash)

	if err != nil {
		return response.Error(401, "Invalid credentials")
	}

	// Verify password
	if !auth.CheckPassword(req.Password, passwordHash) {
		return response.Error(401, "Invalid credentials")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(did, username)
	if err != nil {
		return response.Error(500, "Failed to generate token")
	}

	// Return response
	return response.Success(LoginResponse{
		DID:      did,
		Username: username,
		Token:    token,
	})
}

func main() {
	lambda.Start(handler)
}
