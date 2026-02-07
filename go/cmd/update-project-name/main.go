package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/db"
	"github.com/x-zero/did-login/pkg/response"
)

type UpdateProjectNameRequest struct {
	ProjectName string `json:"project_name"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("=== UpdateProjectName Handler Started ===")
	
	// Handle CORS preflight
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,Authorization",
				"Access-Control-Allow-Methods": "PATCH,OPTIONS",
			},
		}, nil
	}
	
	// Initialize database
	fmt.Println("Initializing database...")
	if err := db.InitDB(); err != nil {
		fmt.Printf("Database init error: %v\n", err)
		return response.Error(500, "Database connection failed")
	}
	fmt.Println("Database initialized successfully")

	// Extract and validate token
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"]
	}

	tokenString, err := auth.ExtractTokenFromHeader(authHeader)
	if err != nil {
		fmt.Printf("Token extraction error: %v\n", err)
		return response.Error(401, "Invalid authorization header")
	}

	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		fmt.Printf("Token validation error: %v\n", err)
		return response.Error(401, "Invalid or expired token")
	}
	fmt.Printf("Token validated, DID: %s\n", claims.DID)

	// Get project ID from path
	projectID := request.PathParameters["id"]
	if projectID == "" {
		return response.Error(400, "Project ID is required")
	}

	// Parse request body
	var req UpdateProjectNameRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		fmt.Printf("JSON parse error: %v\n", err)
		return response.Error(400, "Invalid request body")
	}

	if req.ProjectName == "" {
		return response.Error(400, "Project name is required")
	}

	pool := db.GetPool()

	// Check if user is admin of the project
	var userRole string
	err = pool.QueryRow(ctx,
		"SELECT role FROM user_projects WHERE project_id = $1 AND user_did = $2",
		projectID, claims.DID,
	).Scan(&userRole)

	if err != nil {
		fmt.Printf("User not a member of project: %v\n", err)
		return response.Error(403, "You are not a member of this project")
	}

	if userRole != "admin" {
		return response.Error(403, "Only admins can update project name")
	}

	// Update project name
	_, err = pool.Exec(ctx,
		"UPDATE projects SET project_name = $1, updated_at = NOW() WHERE project_id = $2",
		req.ProjectName, projectID,
	)

	if err != nil {
		fmt.Printf("Database update error: %v\n", err)
		return response.Error(500, "Failed to update project name")
	}

	fmt.Println("Project name updated successfully")
	return response.Success(map[string]string{"message": "Project name updated successfully"})
}

func main() {
	lambda.Start(handler)
}
