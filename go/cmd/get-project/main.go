package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/db"
	"github.com/x-zero/did-login/pkg/models"
	"github.com/x-zero/did-login/pkg/response"
)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("=== GetProject Handler Started ===")
	
	// Handle CORS preflight
	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,Authorization",
				"Access-Control-Allow-Methods": "GET,OPTIONS",
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
	fmt.Printf("Project ID: %s\n", projectID)

	pool := db.GetPool()

	// Check if user is a member of the project
	var userRole string
	err = pool.QueryRow(ctx,
		"SELECT role FROM user_projects WHERE project_id = $1 AND user_did = $2",
		projectID, claims.DID,
	).Scan(&userRole)

	if err != nil {
		fmt.Printf("User not a member of project: %v\n", err)
		return response.Error(403, "You are not a member of this project")
	}

	// Get project details
	var project models.ProjectDetail
	err = pool.QueryRow(ctx,
		"SELECT project_id, project_name, creator_did, created_at, updated_at FROM projects WHERE project_id = $1",
		projectID,
	).Scan(&project.ProjectID, &project.ProjectName, &project.CreatorDID, &project.CreatedAt, &project.UpdatedAt)

	if err != nil {
		fmt.Printf("Project not found: %v\n", err)
		return response.Error(404, "Project not found")
	}

	// Set current user's role
	project.UserRole = userRole

	// Get project members
	rows, err := pool.Query(ctx,
		`SELECT u.did, u.username, u.email, COALESCE(u.profession_tags, '{}'), up.role, up.joined_at
		FROM user_projects up
		JOIN users u ON up.user_did = u.did
		WHERE up.project_id = $1
		ORDER BY up.joined_at ASC`,
		projectID,
	)
	if err != nil {
		fmt.Printf("Failed to get members: %v\n", err)
		return response.Error(500, "Failed to get project members")
	}
	defer rows.Close()

	var members []models.ProjectMember
	for rows.Next() {
		var member models.ProjectMember
		if err := rows.Scan(&member.DID, &member.Username, &member.Email, &member.ProfessionTags, &member.Role, &member.JoinedAt); err != nil {
			fmt.Printf("Row scan error: %v\n", err)
			continue
		}
		member.IsCreator = (member.DID == project.CreatorDID)
		members = append(members, member)
	}

	project.Members = members
	fmt.Printf("Project found with %d members\n", len(members))
	return response.Success(project)
}

func main() {
	lambda.Start(handler)
}
