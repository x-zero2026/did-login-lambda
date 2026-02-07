package main

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/db"
	"github.com/x-zero/did-login/pkg/models"
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

	// Get project ID from path
	projectID := request.PathParameters["id"]
	if projectID == "" {
		return response.Error(400, "Project ID is required")
	}

	// Check if user is member of this project
	pool := db.GetPool()
	var exists bool
	err = pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM user_projects WHERE user_did = $1 AND project_id = $2)",
		claims.DID, projectID,
	).Scan(&exists)

	if err != nil || !exists {
		return response.Error(403, "Access denied to this project")
	}

	// Query apps for this project
	// Include: global apps + project-specific apps - closed apps
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT a.app_id, a.app_name, a.app_description, a.emoji, a.url, a.is_global, u.username, a.created_at
		FROM apps a
		JOIN users u ON a.created_by_did = u.did
		LEFT JOIN app_projects ap ON a.app_id = ap.app_id
		LEFT JOIN project_app_settings pas ON a.app_id = pas.app_id AND pas.project_id = $1
		WHERE (a.is_global = true OR ap.project_id = $1)
		  AND (pas.is_closed IS NULL OR pas.is_closed = false)
		ORDER BY a.created_at ASC
	`, projectID)

	if err != nil {
		return response.Error(500, "Failed to query apps: "+err.Error())
	}
	defer rows.Close()

	var apps []models.AppWithCreator
	for rows.Next() {
		var app models.AppWithCreator
		var createdAt time.Time // Throwaway variable for ORDER BY field
		err := rows.Scan(&app.AppID, &app.AppName, &app.AppDescription, &app.Emoji, &app.URL, &app.IsGlobal, &app.CreatedBy, &createdAt)
		if err != nil {
			return response.Error(500, "Failed to scan app")
		}
		apps = append(apps, app)
	}

	if apps == nil {
		apps = []models.AppWithCreator{}
	}

	return response.Success(apps)
}

func main() {
	lambda.Start(handler)
}
