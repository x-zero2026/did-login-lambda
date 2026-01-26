# Vault Integration Guide

## Overview

This guide explains how the DID Login system integrates with HashiCorp Vault to securely store user mnemonics for XZ Wallet.

---

## Architecture

```
User Registration
  ↓
Generate 1 mnemonic (12 words)
  ↓
Derive 2 addresses:
  1. Solana DID (ed25519) - for authentication
  2. Ethereum address (secp256k1) - for XZ Wallet
  ↓
Store in Database:
  - did (Solana format, 66 chars)
  - eth_address (Ethereum format, 42 chars)
  - email, username, password_hash
  ↓
Store in Vault:
  - Path: secret/data/xz-platform/users/mnemonics/{did}
  - Data: { mnemonic, created_at, version }
```

---

## Vault Path Structure

### Complete Path
```
secret/data/xz-platform/users/mnemonics/{did}
```

### Path Components
- `secret` - KV v2 mount point (configurable via `VAULT_MOUNT_PATH`)
- `data` - KV v2 required prefix
- `xz-platform` - Platform namespace
- `users` - User data category
- `mnemonics` - Mnemonic storage
- `{did}` - User's DID as key (e.g., `0x1234...5678`)

### Example
For user with DID `0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456`:

**Write Path**:
```
secret/data/xz-platform/users/mnemonics/0x1234567890abcdef1234567890abcdef12345678901234567890abcdef123456
```

**Data Structure**:
```json
{
  "data": {
    "mnemonic": "word1 word2 word3 word4 word5 word6 word7 word8 word9 word10 word11 word12",
    "created_at": "2026-01-26T10:00:00Z",
    "version": "1"
  }
}
```

---

## Environment Variables

### Required
```bash
# Vault connection
VAULT_ADDR=https://vault.example.com
VAULT_TOKEN=hvs.xxxxxxxxxxxxx

# Optional (with defaults)
VAULT_MOUNT_PATH=secret                      # KV v2 mount point
VAULT_BASE_PATH=xz-platform/users/mnemonics  # Base path for mnemonics
```

### Local Development (.env)
```bash
# For local Vault instance
VAULT_ADDR=http://127.0.0.1:8200
VAULT_TOKEN=root

# For production
VAULT_ADDR=https://vault.production.com
VAULT_TOKEN=hvs.production_token_here
```

---

## API Endpoints

### 1. Register (Modified)

**Endpoint**: `POST /register`

**Request**:
```json
{
  "email": "user@example.com",
  "username": "testuser",
  "password": "SecurePass123!"
}
```

**Response**:
```json
{
  "did": "0x1234...5678",
  "eth_address": "0xabcd...ef01",
  "username": "testuser",
  "mnemonic": "word1 word2 ... word12",
  "token": "eyJhbGc..."
}
```

**Process**:
1. Generate mnemonic
2. Derive Solana DID (ed25519)
3. Derive Ethereum address (secp256k1, BIP44)
4. Store in database
5. Store mnemonic in Vault
6. Return all info (mnemonic shown once)

---

### 2. Get Ethereum Private Key (New)

**Endpoint**: `GET /get-eth-private-key`

**Headers**:
```
Authorization: Bearer {jwt_token}
```

**Response**:
```json
{
  "did": "0x1234...5678",
  "eth_address": "0xabcd...ef01",
  "private_key": "0x..."
}
```

**Process**:
1. Validate JWT token → extract DID
2. Read mnemonic from Vault
3. Derive Ethereum private key
4. Return private key (temporary use only)

**Security Notes**:
- Private key is derived on-demand
- Never stored in database
- Only exists in memory during request
- Automatically cleared after response

---

## Code Examples

### Store Mnemonic
```go
import "github.com/x-zero/did-login/pkg/vault"

// Create Vault client
vaultClient, err := vault.NewClient()
if err != nil {
    return err
}

// Store mnemonic
err = vaultClient.StoreMnemonic(did, mnemonic)
if err != nil {
    return err
}
```

### Retrieve Mnemonic
```go
// Get mnemonic
mnemonic, err := vaultClient.GetMnemonic(did)
if err != nil {
    return err
}

// Derive Ethereum private key
privateKey, err := auth.DeriveEthereumPrivateKey(mnemonic)
if err != nil {
    return err
}
```

### Derive Ethereum Address
```go
import "github.com/x-zero/did-login/pkg/auth"

// From mnemonic
ethAddress, err := auth.DeriveEthereumAddress(mnemonic)
// Result: 0xabcd...ef01 (42 chars)

// From private key
address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
```

---

## Database Schema

### Users Table (Updated)
```sql
CREATE TABLE users (
    did VARCHAR(66) PRIMARY KEY,           -- Solana format
    eth_address VARCHAR(42),               -- Ethereum format (NEW)
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_eth_address ON users(eth_address);
```

---

## Testing

### 1. Test Vault Connection
```bash
cd did-login-lambda/go
go run cmd/test-vault/main.go
```

### 2. Test Registration
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "Test123!@#"
  }'
```

Expected response includes:
- `did` (66 chars, Solana format)
- `eth_address` (42 chars, Ethereum format)
- `mnemonic` (12 words)

### 3. Test Get Private Key
```bash
# Get token from registration response
TOKEN="eyJhbGc..."

curl -X GET http://localhost:8080/get-eth-private-key \
  -H "Authorization: Bearer $TOKEN"
```

Expected response:
```json
{
  "did": "0x...",
  "eth_address": "0x...",
  "private_key": "0x..."
}
```

### 4. Verify in Vault
```bash
# Using Vault CLI
vault kv get secret/xz-platform/users/mnemonics/0x...

# Expected output:
# ====== Data ======
# Key          Value
# ---          -----
# created_at   2026-01-26T10:00:00Z
# mnemonic     word1 word2 ... word12
# version      1
```

---

## Security Best Practices

### 1. Vault Token Management
- Use short-lived tokens in production
- Rotate tokens regularly
- Use AppRole or AWS IAM auth instead of static tokens

### 2. Mnemonic Handling
- Never log mnemonics
- Only show mnemonic once during registration
- Encourage users to backup mnemonic securely

### 3. Private Key Handling
- Derive on-demand, never store
- Clear from memory after use
- Use HTTPS for all API calls
- Implement rate limiting on `/get-eth-private-key`

### 4. Access Control
- Only authenticated users can access their own private keys
- Implement audit logging for Vault access
- Monitor for suspicious access patterns

---

## Troubleshooting

### Error: "failed to create vault client"
- Check `VAULT_ADDR` is correct
- Verify Vault is running and accessible
- Check network connectivity

### Error: "failed to store mnemonic"
- Verify `VAULT_TOKEN` has write permissions
- Check KV v2 is enabled at `secret/`
- Verify path format is correct

### Error: "mnemonic not found"
- Check DID format is correct
- Verify mnemonic was stored during registration
- Check Vault path: `vault kv get secret/xz-platform/users/mnemonics/{did}`

### Error: "invalid mnemonic"
- Verify mnemonic has exactly 12 words
- Check words are from BIP39 wordlist
- Ensure no extra spaces or special characters

---

## Deployment Checklist

- [ ] Vault is running and accessible
- [ ] KV v2 secrets engine enabled at `secret/`
- [ ] Vault token configured with appropriate permissions
- [ ] Environment variables set in Lambda
- [ ] Database migration completed (`add-eth-address.sql`)
- [ ] Dependencies installed (`go mod tidy`)
- [ ] Lambda functions built and deployed
- [ ] API Gateway routes configured
- [ ] Test registration flow end-to-end
- [ ] Test private key retrieval
- [ ] Verify mnemonics stored in Vault

---

## Next Steps

1. **Deploy to AWS Lambda**
   - Build: `make build` in each cmd directory
   - Upload to Lambda
   - Configure environment variables

2. **Configure API Gateway**
   - Add route: `GET /get-eth-private-key`
   - Enable CORS if needed
   - Set up authorization

3. **Test Integration**
   - Register new user
   - Verify Vault storage
   - Test private key retrieval
   - Integrate with XZ Wallet backend

---

**Document Version**: 1.0  
**Last Updated**: 2026-01-26  
**Status**: Ready for Testing
