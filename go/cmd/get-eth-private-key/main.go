package main

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/response"
	"github.com/x-zero/did-login/pkg/vault"
)

type PrivateKeyResponse struct {
	DID        string `json:"did"`
	EthAddress string `json:"eth_address"`
	PrivateKey string `json:"private_key"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC: %v\n", r)
		}
	}()

	fmt.Println("=== Get Ethereum Private Key Handler Started ===")

	// Get Authorization header
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"] // Try lowercase
	}

	if authHeader == "" {
		return response.Error(401, "Authorization header required")
	}

	// Validate JWT token and extract DID
	claims, err := auth.ValidateToken(authHeader)
	if err != nil {
		fmt.Printf("Token validation error: %v\n", err)
		return response.Error(401, "Invalid or expired token")
	}

	did := claims.DID
	username := claims.Username

	fmt.Printf("Token validated for DID: %s, Username: %s\n", did, username)

	// Initialize Vault client
	vaultClient, err := vault.NewClient()
	if err != nil {
		fmt.Printf("Vault client error: %v\n", err)
		return response.Error(500, "Failed to connect to Vault")
	}

	// Get mnemonic from Vault
	mnemonic, err := vaultClient.GetMnemonic(did)
	if err != nil {
		fmt.Printf("Failed to get mnemonic: %v\n", err)
		return response.Error(500, "Failed to retrieve mnemonic")
	}

	fmt.Println("Mnemonic retrieved from Vault")

	// Derive Ethereum private key
	privateKey, err := auth.DeriveEthereumPrivateKey(mnemonic)
	if err != nil {
		fmt.Printf("Failed to derive private key: %v\n", err)
		return response.Error(500, "Failed to derive private key")
	}

	// Convert private key to hex
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := "0x" + hex.EncodeToString(privateKeyBytes)

	// Get Ethereum address
	ethAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	fmt.Printf("Private key derived successfully for address: %s\n", ethAddress)

	// Return response
	return response.Success(PrivateKeyResponse{
		DID:        did,
		EthAddress: ethAddress,
		PrivateKey: privateKeyHex,
	})
}

func main() {
	lambda.Start(handler)
}
