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
	fmt.Println("=== SearchUsers Handler Started ===")
	
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

	_, err = auth.ValidateToken(tokenString)
	if err != nil {
		fmt.Printf("Token validation error: %v\n", err)
		return response.Error(401, "Invalid or expired token")
	}

	// Get search query
	query := request.QueryStringParameters["q"]
	if query == "" {
		return response.Error(400, "Search query is required")
	}
	fmt.Printf("Search query: %s\n", query)

	// Search users by username or email
	pool := db.GetPool()
	rows, err := pool.Query(ctx,
		`SELECT did, username, email, COALESCE(profession_tags, '{}') 
		FROM users 
		WHERE username ILIKE $1 OR email ILIKE $1 
		LIMIT 10`,
		"%"+query+"%",
	)
	if err != nil {
		fmt.Printf("Database query error: %v\n", err)
		return response.Error(500, "Failed to search users")
	}
	defer rows.Close()

	var users []models.UserSearchResult
	for rows.Next() {
		var user models.UserSearchResult
		if err := rows.Scan(&user.DID, &user.Username, &user.Email, &user.ProfessionTags); err != nil {
			fmt.Printf("Row scan error: %v\n", err)
			continue
		}
		users = append(users, user)
	}

	fmt.Printf("Found %d users\n", len(users))
	return response.Success(users)
}

func main() {
	lambda.Start(handler)
}
