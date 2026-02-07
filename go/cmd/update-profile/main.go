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

type UpdateProfileRequest struct {
	Bio            *string  `json:"bio"`
	ProfessionTags []string `json:"profession_tags"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("=== UpdateProfile Handler Started ===")
	
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
	fmt.Printf("Auth header: %s\n", authHeader)

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

	// Parse request body
	var req UpdateProfileRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		fmt.Printf("JSON parse error: %v\n", err)
		return response.Error(400, "Invalid request body")
	}

	// Validate profession_tags
	if len(req.ProfessionTags) > 5 {
		return response.Error(400, "Maximum 5 profession tags allowed")
	}

	// Validate bio length
	if req.Bio != nil && len(*req.Bio) > 500 {
		return response.Error(400, "Bio must be 500 characters or less")
	}

	// Update user profile
	pool := db.GetPool()
	_, err = pool.Exec(ctx,
		"UPDATE users SET bio = $1, profession_tags = $2, updated_at = NOW() WHERE did = $3",
		req.Bio, req.ProfessionTags, claims.DID,
	)

	if err != nil {
		fmt.Printf("Database update error: %v\n", err)
		return response.Error(500, "Failed to update profile")
	}

	fmt.Println("Profile updated successfully")
	return response.Success(map[string]string{"message": "Profile updated successfully"})
}

func main() {
	lambda.Start(handler)
}
