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

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("=== UpdateProjectMemberRole Handler Started ===")
	
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

	// Get project ID and user DID from path
	projectID := request.PathParameters["id"]
	userDID := request.PathParameters["did"]
	if projectID == "" || userDID == "" {
		return response.Error(400, "Project ID and User DID are required")
	}

	// Parse request body
	var req UpdateRoleRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		fmt.Printf("JSON parse error: %v\n", err)
		return response.Error(400, "Invalid request body")
	}

	if req.Role != "admin" && req.Role != "member" {
		return response.Error(400, "Role must be 'admin' or 'member'")
	}

	pool := db.GetPool()

	// Check if requester is admin of the project
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
		return response.Error(403, "Only admins can update member roles")
	}

	// Check if target user is the project creator
	var creatorDID string
	err = pool.QueryRow(ctx,
		"SELECT creator_did FROM projects WHERE project_id = $1",
		projectID,
	).Scan(&creatorDID)

	if err != nil {
		return response.Error(404, "Project not found")
	}

	if userDID == creatorDID {
		return response.Error(403, "Cannot change project creator's role")
	}

	// Update member role
	result, err := pool.Exec(ctx,
		"UPDATE user_projects SET role = $1 WHERE project_id = $2 AND user_did = $3",
		req.Role, projectID, userDID,
	)

	if err != nil {
		fmt.Printf("Database update error: %v\n", err)
		return response.Error(500, "Failed to update member role")
	}

	if result.RowsAffected() == 0 {
		return response.Error(404, "Member not found in project")
	}

	fmt.Println("Member role updated successfully")
	return response.Success(map[string]string{"message": "Member role updated successfully"})
}

func main() {
	lambda.Start(handler)
}
