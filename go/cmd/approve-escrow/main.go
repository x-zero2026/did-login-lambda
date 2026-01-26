package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/x-zero/did-login/pkg/auth"
	"github.com/x-zero/did-login/pkg/response"
	"github.com/x-zero/did-login/pkg/vault"
)

type ApproveRequest struct {
	TokenAddress   string `json:"token_address"`
	SpenderAddress string `json:"spender_address"`
	Amount         string `json:"amount,omitempty"` // Optional, defaults to max
}

type ApproveResponse struct {
	Success    bool   `json:"success"`
	TxHash     string `json:"tx_hash"`
	EthAddress string `json:"eth_address"`
	Message    string `json:"message"`
}

// Minimal ERC20 ABI for approve function
const erc20ABI = `[{"constant":false,"inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],"name":"approve","outputs":[{"name":"","type":"bool"}],"type":"function"}]`

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC: %v\n", r)
		}
	}()

	fmt.Println("=== Approve Escrow Handler Started ===")

	if request.HTTPMethod == "OPTIONS" {
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Headers: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Headers": "Content-Type,Authorization",
				"Access-Control-Allow-Methods": "POST,OPTIONS",
			},
		}, nil
	}

	// Get Authorization header
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"]
	}

	if authHeader == "" {
		return response.Error(401, "Authorization header required")
	}

	// Extract token from header
	token, err := auth.ExtractTokenFromHeader(authHeader)
	if err != nil {
		fmt.Printf("Failed to extract token: %v\n", err)
		return response.Error(401, "Invalid authorization header format")
	}

	// Validate JWT token and extract DID
	claims, err := auth.ValidateToken(token)
	if err != nil {
		fmt.Printf("Token validation error: %v\n", err)
		return response.Error(401, "Invalid or expired token")
	}

	did := claims.DID
	username := claims.Username
	fmt.Printf("Token validated for DID: %s, Username: %s\n", did, username)

	// Parse request body
	var req ApproveRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return response.Error(400, "Invalid request body")
	}

	if req.TokenAddress == "" || req.SpenderAddress == "" {
		return response.Error(400, "token_address and spender_address are required")
	}

	fmt.Printf("Approve request: Token=%s, Spender=%s\n", req.TokenAddress, req.SpenderAddress)

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

	// Get Ethereum address
	ethAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	fmt.Printf("Derived Ethereum address: %s\n", ethAddress)

	// Connect to Ethereum network
	rpcURL := getEnv("SEPOLIA_RPC_URL", "")
	if rpcURL == "" {
		return response.Error(500, "SEPOLIA_RPC_URL not configured")
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Printf("Failed to connect to Ethereum: %v\n", err)
		return response.Error(500, "Failed to connect to Ethereum network")
	}
	defer client.Close()

	// Get chain ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		fmt.Printf("Failed to get chain ID: %v\n", err)
		return response.Error(500, "Failed to get chain ID")
	}

	// Create transaction auth
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		fmt.Printf("Failed to create auth: %v\n", err)
		return response.Error(500, "Failed to create transaction auth")
	}
	auth.GasLimit = 100000

	// Parse addresses
	tokenAddr := common.HexToAddress(req.TokenAddress)
	spenderAddr := common.HexToAddress(req.SpenderAddress)

	// Determine approval amount (default to max)
	var amount *big.Int
	if req.Amount != "" {
		amount = new(big.Int)
		amount.SetString(req.Amount, 10)
	} else {
		// Max uint256
		amount = new(big.Int)
		amount.SetString("115792089237316195423570985008687907853269984665640564039457584007913129639935", 10)
	}

	// Parse ERC20 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		fmt.Printf("Failed to parse ABI: %v\n", err)
		return response.Error(500, "Failed to parse contract ABI")
	}

	// Create contract instance
	contract := bind.NewBoundContract(tokenAddr, parsedABI, client, client, client)

	// Call approve function
	tx, err := contract.Transact(auth, "approve", spenderAddr, amount)
	if err != nil {
		fmt.Printf("Failed to send approve transaction: %v\n", err)
		return response.Error(500, fmt.Sprintf("Failed to approve: %v", err))
	}

	fmt.Printf("Approve transaction sent: %s\n", tx.Hash().Hex())

	// Wait for transaction to be mined
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		fmt.Printf("Failed to wait for transaction: %v\n", err)
		return response.Error(500, "Transaction sent but confirmation failed")
	}

	if receipt.Status == 0 {
		return response.Error(500, "Approval transaction failed")
	}

	fmt.Printf("Approval confirmed in block %d\n", receipt.BlockNumber.Uint64())

	// Return success response (NO private key!)
	return response.Success(ApproveResponse{
		Success:    true,
		TxHash:     tx.Hash().Hex(),
		EthAddress: ethAddress,
		Message:    "Escrow contract approved successfully",
	})
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	lambda.Start(handler)
}
