package response

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

// Response represents API response structure
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Success returns a successful API Gateway response
func Success(data interface{}) (events.APIGatewayProxyResponse, error) {
	body, err := json.Marshal(Response{
		Success: true,
		Data:    data,
	})
	if err != nil {
		return Error(500, "Failed to marshal response")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
			"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
		},
		Body: string(body),
	}, nil
}

// SuccessWithMessage returns a successful response with message
func SuccessWithMessage(message string) (events.APIGatewayProxyResponse, error) {
	body, err := json.Marshal(Response{
		Success: true,
		Message: message,
	})
	if err != nil {
		return Error(500, "Failed to marshal response")
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
			"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
		},
		Body: string(body),
	}, nil
}

// Error returns an error API Gateway response
func Error(statusCode int, message string) (events.APIGatewayProxyResponse, error) {
	body, _ := json.Marshal(Response{
		Success: false,
		Error:   message,
	})

	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Headers": "Content-Type,Authorization",
			"Access-Control-Allow-Methods": "GET,POST,PUT,DELETE,OPTIONS",
		},
		Body: string(body),
	}, nil
}
