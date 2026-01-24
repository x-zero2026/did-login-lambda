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

type UpdateAppRequest struct {
	AppName        string `json:"app_name"`
	AppDescription string `json:"app_description"`
	Emoji          string `json:"emoji"`
	URL            string `json:"url"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Initialize database
	if err := db.InitDB(); err != nil {
		return response.Error(500, "Database connection failed")
	}

	// Extract and validate token
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"]
	}

	tokenString, err := auth.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return response.Error(401, "Invalid authorization header")
	}

	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		return response.Error(401, "Invalid or expired token")
	}

	// Get app ID from path
	appID := request.PathParameters["id"]
	if appID == "" {
		return response.Error(400, "App ID is required")
	}

	// Parse request body
	var req UpdateAppRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.Error(400, "Invalid request body")
	}

	// Validate input
	if req.AppName == "" || req.Emoji == "" || req.URL == "" {
		return response.Error(400, "App name, emoji, and URL are required")
	}

	// Check if user is the creator of this app
	pool := db.GetPool()
	var creatorDID string
	err = pool.QueryRow(ctx,
		"SELECT created_by_did FROM apps WHERE app_id = $1",
		appID,
	).Scan(&creatorDID)

	if err != nil {
		return response.Error(404, "App not found")
	}

	if creatorDID != claims.DID {
		return response.Error(403, "Only the creator can update this app")
	}

	// Update app
	_, err = pool.Exec(ctx, `
		UPDATE apps
		SET app_name = $1, app_description = $2, emoji = $3, url = $4, updated_at = CURRENT_TIMESTAMP
		WHERE app_id = $5
	`, req.AppName, req.AppDescription, req.Emoji, req.URL, appID)

	if err != nil {
		return response.Error(500, "Failed to update app")
	}

	return response.SuccessWithMessage("App updated successfully")
}

func main() {
	lambda.Start(handler)
}
