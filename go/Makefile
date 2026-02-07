.PHONY: build-RegisterFunction build-LoginFunction build-RecoverFunction build-ResetPasswordFunction build-GetProfileFunction build-UpdateProfileFunction build-SearchUsersFunction build-GetProjectFunction build-UpdateProjectNameFunction build-AddProjectMemberFunction build-RemoveProjectMemberFunction build-UpdateProjectMemberRoleFunction build-ListProjectsFunction build-UpdateProjectFunction build-ListAppsFunction build-CreateAppFunction build-UpdateAppFunction build-DeleteAppFunction build-SetAppGlobalFunction build-ToggleAppFunction build-ApproveEscrowFunction

build-RegisterFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/register/main.go

build-LoginFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/login/main.go

build-RecoverFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/recover/main.go

build-ResetPasswordFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/reset-password/main.go

build-GetProfileFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/get-profile/main.go

build-UpdateProfileFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/update-profile/main.go

build-SearchUsersFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/search-users/main.go

build-GetProjectFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/get-project/main.go

build-UpdateProjectNameFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/update-project-name/main.go

build-AddProjectMemberFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/add-project-member/main.go

build-RemoveProjectMemberFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/remove-project-member/main.go

build-UpdateProjectMemberRoleFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/update-project-member-role/main.go

build-ListProjectsFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/list-projects/main.go

build-UpdateProjectFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/update-project/main.go

build-ListAppsFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/list-apps/main.go

build-CreateAppFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/create-app/main.go

build-UpdateAppFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/update-app/main.go

build-DeleteAppFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/delete-app/main.go

build-SetAppGlobalFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/set-app-global/main.go

build-ToggleAppFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/toggle-app/main.go

build-ApproveEscrowFunction:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o $(ARTIFACTS_DIR)/bootstrap ./cmd/approve-escrow/main.go
