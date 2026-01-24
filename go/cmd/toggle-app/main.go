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

type ToggleAppRequest struct {
	IsClosed bool `json:"is_closed"`
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

	// Get project ID and app ID from path
	projectID := request.PathParameters["project_id"]
	appID := request.PathParameters["app_id"]
	if projectID == "" || appID == "" {
		return response.Error(400, "Project ID and App ID are required")
	}

	// Parse request body
	var req ToggleAppRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.Error(400, "Invalid request body")
	}

	// Check if user is admin of this project
	pool := db.GetPool()
	var role string
	err = pool.QueryRow(ctx,
		"SELECT role FROM user_projects WHERE user_did = $1 AND project_id = $2",
		claims.DID, projectID,
	).Scan(&role)

	if err != nil {
		return response.Error(404, "Project not found or access denied")
	}

	if role != "admin" {
		return response.Error(403, "Only admins can toggle app status")
	}

	// Upsert project_app_settings
	_, err = pool.Exec(ctx, `
		INSERT INTO project_app_settings (project_id, app_id, is_closed)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id, app_id)
		DO UPDATE SET is_closed = $3, updated_at = CURRENT_TIMESTAMP
	`, projectID, appID, req.IsClosed)

	if err != nil {
		return response.Error(500, "Failed to update app status")
	}

	message := "App enabled for this project"
	if req.IsClosed {
		message = "App disabled for this project"
	}

	return response.SuccessWithMessage(message)
}

func main() {
	lambda.Start(handler)
}
