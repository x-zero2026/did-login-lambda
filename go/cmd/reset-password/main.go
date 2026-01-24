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

type ResetPasswordRequest struct {
	Mnemonic    string `json:"mnemonic"`
	NewPassword string `json:"new_password"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize database
	if err := db.InitDB(); err != nil {
		return response.Error(500, "Database connection failed")
	}

	// Parse request body
	var req ResetPasswordRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.Error(400, "Invalid request body")
	}

	// Validate input
	if req.Mnemonic == "" || req.NewPassword == "" {
		return response.Error(400, "Mnemonic and new password are required")
	}

	// Generate DID from mnemonic
	did, err := auth.MnemonicToDID(req.Mnemonic)
	if err != nil {
		return response.Error(400, "Invalid mnemonic")
	}

	// Check if user exists
	pool := db.GetPool()
	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE did = $1)", did).Scan(&exists)
	if err != nil || !exists {
		return response.Error(404, "User not found")
	}

	// Hash new password
	passwordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return response.Error(500, "Failed to hash password")
	}

	// Update password
	_, err = pool.Exec(ctx,
		"UPDATE users SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE did = $2",
		passwordHash, did,
	)
	if err != nil {
		return response.Error(500, "Failed to update password")
	}

	return response.SuccessWithMessage("Password reset successfully")
}

func main() {
	lambda.Start(handler)
}
