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

type AddMemberRequest struct {
	UserDID string `json:"user_did"`
	Role    string `json:"role"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("=== AddProjectMember Handler Started ===")
	
	// Handle CORS preflight
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,Authorization",
				"Access-Control-Allow-Methods": "POST,OPTIONS",
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
	var req AddMemberRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		fmt.Printf("JSON parse error: %v\n", err)
		return response.Error(400, "Invalid request body")
	}

	if req.UserDID == "" {
		return response.Error(400, "User DID is required")
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
		return response.Error(403, "Only admins can add members")
	}

	// Check if user exists
	var exists bool
	err = pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE did = $1)",
		req.UserDID,
	).Scan(&exists)

	if err != nil || !exists {
		return response.Error(404, "User not found")
	}

	// Check if user is already a member
	err = pool.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM user_projects WHERE project_id = $1 AND user_did = $2)",
		projectID, req.UserDID,
	).Scan(&exists)

	if err == nil && exists {
		return response.Error(400, "User is already a member of this project")
	}

	// Add member to project
	_, err = pool.Exec(ctx,
		"INSERT INTO user_projects (user_did, project_id, role, joined_at) VALUES ($1, $2, $3, NOW())",
		req.UserDID, projectID, req.Role,
	)

	if err != nil {
		fmt.Printf("Database insert error: %v\n", err)
		return response.Error(500, "Failed to add member")
	}

	fmt.Println("Member added successfully")
	return response.Success(map[string]string{"message": "Member added successfully"})
}

func main() {
	lambda.Start(handler)
}
