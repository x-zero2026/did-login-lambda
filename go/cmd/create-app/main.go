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

type CreateAppRequest struct {
	ProjectID      string `json:"project_id"`
	AppName        string `json:"app_name"`
	AppDescription string `json:"app_description"`
	Emoji          string `json:"emoji"`
	URL            string `json:"url"`
}

type CreateAppResponse struct {
	AppID   string `json:"app_id"`
	AppName string `json:"app_name"`
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

	// Parse request body
	var req CreateAppRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.Error(400, "Invalid request body")
	}

	// Validate input
	if req.ProjectID == "" || req.AppName == "" || req.Emoji == "" || req.URL == "" {
		return response.Error(400, "Project ID, app name, emoji, and URL are required")
	}

	// Check if user is admin of this project
	pool := db.GetPool()
	var role string
	err = pool.QueryRow(ctx,
		"SELECT role FROM user_projects WHERE user_did = $1 AND project_id = $2",
		claims.DID, req.ProjectID,
	).Scan(&role)

	if err != nil {
		return response.Error(404, "Project not found or access denied")
	}

	if role != "admin" {
		return response.Error(403, "Only admins can create apps")
	}

	// Start transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return response.Error(500, "Failed to start transaction")
	}
	defer tx.Rollback(ctx)

	// Create app (is_global = false by default)
	var appID string
	err = tx.QueryRow(ctx, `
		INSERT INTO apps (app_name, app_description, emoji, url, is_global, created_by_did)
		VALUES ($1, $2, $3, $4, false, $5)
		RETURNING app_id
	`, req.AppName, req.AppDescription, req.Emoji, req.URL, claims.DID).Scan(&appID)

	if err != nil {
		return response.Error(500, "Failed to create app")
	}

	// Link app to project
	_, err = tx.Exec(ctx,
		"INSERT INTO app_projects (app_id, project_id) VALUES ($1, $2)",
		appID, req.ProjectID,
	)
	if err != nil {
		return response.Error(500, "Failed to link app to project")
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return response.Error(500, "Failed to commit transaction")
	}

	return response.Success(CreateAppResponse{
		AppID:   appID,
		AppName: req.AppName,
	})
}

func main() {
	lambda.Start(handler)
}
