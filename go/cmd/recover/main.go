package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/db"
	"github.com/x-zero/did-login/pkg/response"
)

type RecoverRequest struct {
	Mnemonic string `json:"mnemonic"`
}

type RecoverResponse struct {
	DID      string `json:"did"`
	Username string `json:"username"`
	Email    string `json:"email"` // Masked email
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	local := parts[0]
	domain := parts[1]

	if len(local) <= 2 {
		return local[0:1] + "***@" + domain
	}

	return local[0:1] + "***" + local[len(local)-1:] + "@" + domain
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize database
	if err := db.InitDB(); err != nil {
		return response.Error(500, "Database connection failed")
	}

	// Parse request body
	var req RecoverRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.Error(400, "Invalid request body")
	}

	// Validate input
	if req.Mnemonic == "" {
		return response.Error(400, "Mnemonic is required")
	}

	// Generate DID from mnemonic
	did, err := auth.MnemonicToDID(req.Mnemonic)
	if err != nil {
		return response.Error(400, "Invalid mnemonic")
	}

	// Query user by DID
	pool := db.GetPool()
	var username, email string
	err = pool.QueryRow(ctx,
		"SELECT username, email FROM users WHERE did = $1",
		did,
	).Scan(&username, &email)

	if err != nil {
		return response.Error(404, "User not found with this mnemonic")
	}

	// Return response with masked email
	return response.Success(RecoverResponse{
		DID:      did,
		Username: username,
		Email:    maskEmail(email),
	})
}

func main() {
	lambda.Start(handler)
}
