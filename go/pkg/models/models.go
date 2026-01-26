package models

import "time"

// User represents a user in the system
type User struct {
	DID          string    `json:"did"`
	Email        string    `json:"email"`
	Username     string    `json:"username"`
	EthAddress   string    `json:"eth_address"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Project represents a project
type Project struct {
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	CreatorDID  string    `json:"creator_did"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserProject represents user-project relationship
type UserProject struct {
	UserDID   string    `json:"user_did"`
	ProjectID string    `json:"project_id"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// ProjectWithRole includes project info and user's role
type ProjectWithRole struct {
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joined_at"`
	IsDefault   bool      `json:"is_default"`
}

// App represents an application
type App struct {
	AppID          string    `json:"app_id"`
	AppName        string    `json:"app_name"`
	AppDescription string    `json:"app_description"`
	Emoji          string    `json:"emoji"`
	URL            string    `json:"url"`
	IsGlobal       bool      `json:"is_global"`
	CreatedByDID   string    `json:"created_by_did"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// AppWithCreator includes app info and creator username
type AppWithCreator struct {
	AppID          string `json:"app_id"`
	AppName        string `json:"app_name"`
	AppDescription string `json:"app_description"`
	Emoji          string `json:"emoji"`
	URL            string `json:"url"`
	IsGlobal       bool   `json:"is_global"`
	CreatedBy      string `json:"created_by"`
}
