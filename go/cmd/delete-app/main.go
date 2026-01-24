package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/db"
	"github.com/x-zero/did-login/pkg/response"
)

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
		return response.Error(403, "Only the creator can delete this app")
	}

	// Delete app (cascade will delete app_projects and project_app_settings)
	_, err = pool.Exec(ctx, "DELETE FROM apps WHERE app_id = $1", appID)
	if err != nil {
		return response.Error(500, "Failed to delete app")
	}

	return response.SuccessWithMessage("App deleted successfully")
}

func main() {
	lambda.Start(handler)
}
