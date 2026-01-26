package auth

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/tyler-smith/go-bip39"
)

// DeriveEthereumAddress derives Ethereum address from existing mnemonic
// Uses BIP44 path: m/44'/60'/0'/0/0
func DeriveEthereumAddress(mnemonic string) (string, error) {
	privateKey, err := DeriveEthereumPrivateKey(mnemonic)
	if err != nil {
		return "", err
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return address.Hex(), nil
}

// DeriveEthereumPrivateKey derives Ethereum private key from mnemonic
// Uses BIP44 path: m/44'/60'/0'/0/0
func DeriveEthereumPrivateKey(mnemonic string) (*ecdsa.PrivateKey, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid mnemonic")
	}

	// 生成 seed
	seed := bip39.NewSeed(mnemonic, "")

	// BIP44 derivation: m/44'/60'/0'/0/0
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create master key: %w", err)
	}

	// m/44'
	purposeKey, err := masterKey.Derive(hdkeychain.HardenedKeyStart + 44)
	if err != nil {
		return nil, fmt.Errorf("failed to derive purpose key: %w", err)
	}

	// m/44'/60'
	coinKey, err := purposeKey.Derive(hdkeychain.HardenedKeyStart + 60)
	if err != nil {
		return nil, fmt.Errorf("failed to derive coin key: %w", err)
	}

	// m/44'/60'/0'
	accountKey, err := coinKey.Derive(hdkeychain.HardenedKeyStart + 0)
	if err != nil {
		return nil, fmt.Errorf("failed to derive account key: %w", err)
	}

	// m/44'/60'/0'/0
	changeKey, err := accountKey.Derive(0)
	if err != nil {
		return nil, fmt.Errorf("failed to derive change key: %w", err)
	}

	// m/44'/60'/0'/0/0
	addressKey, err := changeKey.Derive(0)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address key: %w", err)
	}

	// 获取私钥
	privateKeyECDSA, err := addressKey.ECPrivKey()
	if err != nil {
		return nil, fmt.Errorf("failed to get private key: %w", err)
	}

	return privateKeyECDSA.ToECDSA(), nil
}
