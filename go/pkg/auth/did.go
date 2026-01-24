package auth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/tyler-smith/go-bip39"
)

// GenerateMnemonic generates a 12-word BIP39 mnemonic
func GenerateMnemonic() (string, error) {
	// Generate 128 bits of entropy (16 bytes) for 12 words
	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		return "", fmt.Errorf("failed to generate entropy: %w", err)
	}

	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		return "", fmt.Errorf("failed to generate mnemonic: %w", err)
	}

	return mnemonic, nil
}

// MnemonicToDID converts a mnemonic to a DID (0x-prefixed address)
func MnemonicToDID(mnemonic string) (string, error) {
	// Validate mnemonic
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", fmt.Errorf("invalid mnemonic")
	}

	// Generate seed from mnemonic (64 bytes)
	seed := bip39.NewSeed(mnemonic, "")

	// Use first 32 bytes as ed25519 private key
	privateKey := ed25519.NewKeyFromSeed(seed[:32])

	// Generate public key
	publicKey := privateKey.Public().(ed25519.PublicKey)

	// Convert public key to hex with 0x prefix
	did := "0x" + hex.EncodeToString(publicKey)

	return did, nil
}

// GenerateDIDAndMnemonic generates both a mnemonic and its corresponding DID
func GenerateDIDAndMnemonic() (did string, mnemonic string, err error) {
	mnemonic, err = GenerateMnemonic()
	if err != nil {
		return "", "", err
	}

	did, err = MnemonicToDID(mnemonic)
	if err != nil {
		return "", "", err
	}

	return did, mnemonic, nil
}

// GenerateRandomDID generates a random DID without mnemonic (for testing)
func GenerateRandomDID() (string, error) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}

	did := "0x" + hex.EncodeToString(publicKey)
	return did, nil
}
