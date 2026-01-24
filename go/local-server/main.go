package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joho/godotenv"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/db"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Success  bool   `json:"success"`
	DID      string `json:"did,omitempty"`
	Username string `json:"username,omitempty"`
	Mnemonic string `json:"mnemonic,omitempty"`
	Token    string `json:"token,omitempty"`
	Error    string `json:"error,omitempty"`
}

var usernameRegex = regexp.MustCompile(`^[a-zA-Z]([a-zA-Z-]*[a-zA-Z])?$`)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Method not allowed"})
		return
	}

	// Read body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Failed to read request body"})
		return
	}

	// Parse request
	var req RegisterRequest
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Invalid request body"})
		return
	}

	// Validate input
	if req.Email == "" || req.Username == "" || req.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Email, username, and password are required"})
		return
	}

	// Validate username format
	if len(req.Username) > 255 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Username must be 255 characters or less"})
		return
	}
	if !usernameRegex.MatchString(req.Username) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Username must contain only letters and hyphens, cannot start or end with hyphen"})
		return
	}

	// Validate email format
	if !strings.Contains(req.Email, "@") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Invalid email format"})
		return
	}

	// Check if email or username already exists
	pool := db.GetPool()
	ctx := context.Background()
	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1 OR username = $2)", req.Email, req.Username).Scan(&exists)
	if err != nil {
		log.Printf("Database query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Database query failed"})
		return
	}
	if exists {
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Email or username already exists"})
		return
	}

	// Generate DID and mnemonic
	did, mnemonic, err := auth.GenerateDIDAndMnemonic()
	if err != nil {
		log.Printf("Failed to generate DID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Failed to generate DID"})
		return
	}

	// Hash password
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Failed to hash password"})
		return
	}

	// Start transaction
	tx, err := pool.Begin(ctx)
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Failed to start transaction"})
		return
	}
	defer tx.Rollback(ctx)

	// Insert user
	_, err = tx.Exec(ctx,
		"INSERT INTO users (did, email, username, password_hash) VALUES ($1, $2, $3, $4)",
		did, req.Email, req.Username, passwordHash,
	)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Failed to create user"})
		return
	}

	// Create default project
	var projectID string
	err = tx.QueryRow(ctx,
		"INSERT INTO projects (project_name, creator_did) VALUES ($1, $2) RETURNING project_id",
		req.Username, did,
	).Scan(&projectID)
	if err != nil {
		log.Printf("Failed to create project: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Failed to create default project"})
		return
	}

	// Add user to project as admin
	_, err = tx.Exec(ctx,
		"INSERT INTO user_projects (user_did, project_id, role) VALUES ($1, $2, $3)",
		did, projectID, "admin",
	)
	if err != nil {
		log.Printf("Failed to add user to project: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Failed to add user to project"})
		return
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		log.Printf("Failed to commit transaction: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Failed to commit transaction"})
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(did, req.Username)
	if err != nil {
		log.Printf("Failed to generate token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RegisterResponse{Success: false, Error: "Failed to generate token"})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(RegisterResponse{
		Success:  true,
		DID:      did,
		Username: req.Username,
		Mnemonic: mnemonic,
		Token:    token,
	})
}

// Login handler
func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}

	body, _ := io.ReadAll(r.Body)
	var req struct {
		EmailOrUsername string `json:"emailOrUsername"`
		Password        string `json:"password"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid request body"})
		return
	}

	// Query user
	pool := db.GetPool()
	ctx := context.Background()
	var did, username, passwordHash string
	err := pool.QueryRow(ctx,
		"SELECT did, username, password_hash FROM users WHERE email = $1 OR username = $1",
		req.EmailOrUsername,
	).Scan(&did, &username, &passwordHash)
	
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid credentials"})
		return
	}

	// Verify password
	if !auth.CheckPassword(req.Password, passwordHash) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid credentials"})
		return
	}

	// Generate token
	token, err := auth.GenerateToken(did, username)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to generate token"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"did":      did,
		"username": username,
		"token":    token,
	})
}

// Get profile handler
func getProfileHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}

	// Get token from header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No authorization token"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid token"})
		return
	}

	// Get user info
	pool := db.GetPool()
	ctx := context.Background()
	var email, username string
	err = pool.QueryRow(ctx,
		"SELECT email, username FROM users WHERE did = $1",
		claims.DID,
	).Scan(&email, &username)
	
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "User not found"})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"did":      claims.DID,
		"email":    email,
		"username": username,
	})
}

// Get projects handler
func getProjectsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}

	// Get token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No authorization token"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := auth.ValidateToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid token"})
		return
	}

	// Get user's projects
	pool := db.GetPool()
	ctx := context.Background()
	rows, err := pool.Query(ctx, `
		SELECT p.project_id, p.project_name, p.creator_did, p.created_at, up.role, up.joined_at
		FROM projects p
		JOIN user_projects up ON p.project_id = up.project_id
		WHERE up.user_did = $1
		ORDER BY up.joined_at DESC
	`, claims.DID)
	
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to query projects"})
		return
	}
	defer rows.Close()

	projects := []map[string]interface{}{}
	for rows.Next() {
		var projectID, projectName, creatorDID, role string
		var createdAt, joinedAt interface{}
		rows.Scan(&projectID, &projectName, &creatorDID, &createdAt, &role, &joinedAt)
		projects = append(projects, map[string]interface{}{
			"project_id":   projectID,
			"project_name": projectName,
			"creator_did":  creatorDID,
			"role":         role,
			"created_at":   createdAt,
			"joined_at":    joinedAt,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"projects": projects,
	})
}

// Get project apps handler
func getProjectAppsHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Method not allowed"})
		return
	}

	// Extract project_id from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid URL"})
		return
	}
	projectID := parts[3]

	// Get token
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "No authorization token"})
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	_, err := auth.ValidateToken(tokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Invalid token"})
		return
	}

	// Get apps for project
	pool := db.GetPool()
	ctx := context.Background()
	rows, err := pool.Query(ctx, `
		SELECT DISTINCT a.app_id, a.app_name, a.app_description, a.emoji, a.url, a.created_by_did, a.created_at
		FROM apps a
		LEFT JOIN app_projects ap ON a.app_id = ap.app_id
		LEFT JOIN project_app_settings pas ON a.app_id = pas.app_id AND pas.project_id = $1
		WHERE (ap.project_id = $1 OR ap.project_id IS NULL)
		AND (pas.is_closed IS NULL OR pas.is_closed = false)
		ORDER BY a.created_at DESC
	`, projectID)
	
	if err != nil {
		log.Printf("Query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": "Failed to query apps"})
		return
	}
	defer rows.Close()

	apps := []map[string]interface{}{}
	for rows.Next() {
		var appID, appName, appDescription, emoji, url, createdByDID string
		var createdAt interface{}
		rows.Scan(&appID, &appName, &appDescription, &emoji, &url, &createdByDID, &createdAt)
		apps = append(apps, map[string]interface{}{
			"app_id":          appID,
			"app_name":        appName,
			"app_description": appDescription,
			"emoji":           emoji,
			"url":             url,
			"created_by_did":  createdByDID,
			"created_at":      createdAt,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"apps":    apps,
	})
}

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Initialize database
	log.Println("Initializing database connection...")
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	log.Println("Database connected successfully!")

	// Setup routes
	http.HandleFunc("/api/auth/register", registerHandler)
	http.HandleFunc("/api/auth/login", loginHandler)
	http.HandleFunc("/api/user/profile", getProfileHandler)
	http.HandleFunc("/api/projects", getProjectsHandler)
	http.HandleFunc("/api/projects/", getProjectAppsHandler) // Handles /api/projects/{id}/apps

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("🚀 Local server starting on http://localhost:%s", port)
	log.Printf("📍 API endpoints:")
	log.Printf("   POST   /api/auth/register")
	log.Printf("   POST   /api/auth/login")
	log.Printf("   GET    /api/user/profile")
	log.Printf("   GET    /api/projects")
	log.Printf("   GET    /api/projects/{id}/apps")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
