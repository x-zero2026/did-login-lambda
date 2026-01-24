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

type SetGlobalRequest struct {
	IsGlobal bool `json:"is_global"`
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
	var req SetGlobalRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.Error(400, "Invalid request body")
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
		return response.Error(403, "Only the creator can change global status")
	}

	// Start transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return response.Error(500, "Failed to start transaction")
	}
	defer tx.Rollback(ctx)

	// Update is_global
	_, err = tx.Exec(ctx,
		"UPDATE apps SET is_global = $1, updated_at = CURRENT_TIMESTAMP WHERE app_id = $2",
		req.IsGlobal, appID,
	)
	if err != nil {
		return response.Error(500, "Failed to update app")
	}

	// If setting to global, remove project associations
	if req.IsGlobal {
		_, err = tx.Exec(ctx, "DELETE FROM app_projects WHERE app_id = $1", appID)
		if err != nil {
			return response.Error(500, "Failed to remove project associations")
		}
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return response.Error(500, "Failed to commit transaction")
	}

	message := "App set to project-specific"
	if req.IsGlobal {
		message = "App set to global"
	}

	return response.SuccessWithMessage(message)
}

func main() {
	lambda.Start(handler)
}
