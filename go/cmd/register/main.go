package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/db"
	"github.com/x-zero/did-login/pkg/response"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	DID      string `json:"did"`
	Username string `json:"username"`
	Mnemonic string `json:"mnemonic"`
	Token    string `json:"token"`
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z-]*[a-zA-Z])?$`)

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC: %v\n", r)
		}
	}()
	
	fmt.Println("=== Register Handler Started ===")
	fmt.Printf("Request Body: %s\n", request.Body)
	
	// Initialize database (singleton, only runs once)
	fmt.Println("Initializing database...")
	if err := db.InitDB(); err != nil {
		fmt.Printf("Database init error: %v\n", err)
		return response.Error(500, fmt.Sprintf("Database connection failed: %v", err))
	}
	fmt.Println("Database initialized successfully")

	// Parse request body
	var req RegisterRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		return response.Error(400, "Invalid request body")
	}
	fmt.Printf("Parsed request: email=%s, username=%s\n", req.Email, req.Username)

	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		return response.Error(400, "Email, username, and password are required")
	}

	// Validate username format
	if len(req.Username) > 255 {
		return response.Error(400, "Username must be 255 characters or less")
	}
	if !usernameRegex.MatchString(req.Username) {
		return response.Error(400, "Username must contain only letters and hyphens, cannot start or end with hyphen")
	}

	// Validate email format (basic)
	if !strings.Contains(req.Email, "@") {
		return response.Error(400, "Invalid email format")
	}

	// Check if email or username already exists
	pool := db.GetPool()
	var exists bool
	err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)", req.Email, req.Username).Scan(&exists)
	if err != nil {
		fmt.Printf("Database query error: %v\n", err)
		return response.Error(500, fmt.Sprintf("Database query failed: %v", err))
	}
	if exists {
		return response.Error(409, "Email or username already exists")
	}

	// Generate DID and mnemonic
	did, mnemonic, err := auth.GenerateDIDAndMnemonic()
	if err != nil {
		return response.Error(500, "Failed to generate DID")
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		return response.Error(500, "Failed to hash password")
	}

	// Start transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		return response.Error(500, "Failed to start transaction")
	}
	defer tx.Rollback(ctx)

	// Insert user
	_, err = tx.Exec(ctx,
		"INSERT INTO users (did, email, username, password_hash) VALUES ($1, $2, $3, $4)",
		did, req.Email, req.Username, passwordHash,
	)
	if err != nil {
		return response.Error(500, fmt.Sprintf("Failed to create user: %v", err))
	}

	// Create default project (project name = username)
	var projectID string
	err = tx.QueryRow(ctx,
		"INSERT INTO projects (project_name, creator_did) VALUES ($1, $2) RETURNING project_id",
		req.Username, did,
	).Scan(&projectID)
	if err != nil {
		return response.Error(500, "Failed to create default project")
	}

	// Add user to project as admin
	_, err = tx.Exec(ctx,
		"INSERT INTO user_projects (user_did, project_id, role) VALUES ($1, $2, $3)",
		did, projectID, "admin",
	)
	if err != nil {
		return response.Error(500, "Failed to add user to project")
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return response.Error(500, "Failed to commit transaction")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(did, req.Username)
	if err != nil {
		return response.Error(500, "Failed to generate token")
	}

	// Return response
	return response.Success(RegisterResponse{
		DID:      did,
		Username: req.Username,
		Mnemonic: mnemonic,
		Token:    token,
	})
}

func main() {
	lambda.Start(handler)
}
