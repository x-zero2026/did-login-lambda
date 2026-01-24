package main

import (
	"context"

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

	// Query user's projects
	pool := db.GetPool()
	rows, err := pool.Query(ctx, `
		SELECT p.project_id, p.project_name, up.role, up.joined_at
		FROM user_projects up
		JOIN projects p ON up.project_id = p.project_id
		WHERE up.user_did = $1
		ORDER BY up.joined_at DESC
	`, claims.DID)

	if err != nil {
		return response.Error(500, "Failed to query projects")
	}
	defer rows.Close()

	var projects []models.ProjectWithRole
	for rows.Next() {
		var project models.ProjectWithRole
		err := rows.Scan(&project.ProjectID, &project.ProjectName, &project.Role, &project.JoinedAt)
		if err != nil {
			return response.Error(500, "Failed to scan project")
		}

		// Check if this is the default project (project_name == username)
		project.IsDefault = (project.ProjectName == claims.Username)

		projects = append(projects, project)
	}

	if projects == nil {
		projects = []models.ProjectWithRole{}
	}

	return response.Success(projects)
}

func main() {
	lambda.Start(handler)
}
