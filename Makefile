.PHONY: build clean

build:
	@echo "Building Lambda functions..."
	@cd go/cmd/register && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/login && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/recover && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/reset-password && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/get-profile && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/list-projects && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/update-project && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/list-apps && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/create-app && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/update-app && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/delete-app && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/set-app-global && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@cd go/cmd/toggle-app && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	@echo "✅ Build complete!"

clean:
	@echo "Cleaning build artifacts..."
	@find go/cmd -name bootstrap -delete
	@rm -rf .aws-sam
	@echo "✅ Clean complete!"
