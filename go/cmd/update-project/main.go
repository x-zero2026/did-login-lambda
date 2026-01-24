package main

import (
	"context"
	"encoding/json"
	"regexp"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/db"
	"github.com/x-zero/did-login/pkg/response"
)

type UpdateProjectRequest struct {
	ProjectName string `json:"project_name"`
}

var projectNameRegex = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z-]*[a-zA-Z])?$`)

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

	// Parse request body
	var req UpdateProjectRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.Error(400, "Invalid request body")
	}

	// Validate project name
	if req.ProjectName == "" {
		return response.Error(400, "Project name is required")
	}
	if len(req.ProjectName) > 255 {
		return response.Error(400, "Project name must be 255 characters or less")
	}
	if !projectNameRegex.MatchString(req.ProjectName) {
		return response.Error(400, "Project name must contain only letters and hyphens, cannot start or end with hyphen")
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
		return response.Error(403, "Only admins can update project name")
	}

	// Update project name
	_, err = pool.Exec(ctx,
		"UPDATE projects SET project_name = $1, updated_at = CURRENT_TIMESTAMP WHERE project_id = $2",
		req.ProjectName, projectID,
	)
	if err != nil {
		return response.Error(500, "Failed to update project")
	}

	return response.SuccessWithMessage("Project name updated successfully")
}

func main() {
	lambda.Start(handler)
}
