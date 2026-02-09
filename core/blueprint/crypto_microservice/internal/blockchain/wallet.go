package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Wallet represents a cryptocurrency wallet
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
	Address    string
	Currency   string
}

// GenerateWallet generates a new cryptocurrency wallet
func GenerateWallet(currency string) (*Wallet, error) {
	// Generate ECDSA private key
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Get public key
	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)

	// Generate address based on currency
	address := generateAddress(publicKey, currency)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
		Currency:   currency,
	}, nil
}

// generateAddress generates a blockchain address from public key
func generateAddress(publicKey []byte, currency string) string {
	hash := sha256.Sum256(publicKey)

	switch currency {
	case "BTC":
		// Bitcoin testnet addresses start with 'tb1' or 'n'/'m'
		return "tb1q" + hex.EncodeToString(hash[:20])
	case "ETH", "USDT":
		// Ethereum addresses are 0x + 40 hex characters
		return "0x" + hex.EncodeToString(hash[:20])
	default:
		return hex.EncodeToString(hash[:])
	}
}

// SignMessage signs a message with the wallet's private key
func (w *Wallet) SignMessage(message []byte) (string, error) {
	hash := sha256.Sum256(message)

	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature), nil
}

// VerifySignature verifies a signature against a message
func VerifySignature(publicKey []byte, message []byte, signature string) (bool, error) {
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("invalid signature format: %w", err)
	}

	if len(sigBytes) != 64 {
		return false, fmt.Errorf("invalid signature length")
	}

	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	hash := sha256.Sum256(message)

	// Reconstruct public key
	x := new(big.Int).SetBytes(publicKey[:32])
	y := new(big.Int).SetBytes(publicKey[32:])

	pubKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return ecdsa.Verify(pubKey, hash[:], r, s), nil
}

// GetPrivateKeyHex returns the private key as a hex string
func (w *Wallet) GetPrivateKeyHex() string {
	return hex.EncodeToString(w.PrivateKey.D.Bytes())
}

// GetPublicKeyHex returns the public key as a hex string
func (w *Wallet) GetPublicKeyHex() string {
	return hex.EncodeToString(w.PublicKey)
}

// WalletFromPrivateKey reconstructs a wallet from a private key hex string
func WalletFromPrivateKey(privateKeyHex string, currency string) (*Wallet, error) {
	keyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key format: %w", err)
	}

	privateKey := new(ecdsa.PrivateKey)
	privateKey.D = new(big.Int).SetBytes(keyBytes)
	privateKey.PublicKey.Curve = elliptic.P256()
	privateKey.PublicKey.X, privateKey.PublicKey.Y = privateKey.PublicKey.Curve.ScalarBaseMult(keyBytes)

	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	address := generateAddress(publicKey, currency)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
		Currency:   currency,
	}, nil
}

// ValidateAddress validates a cryptocurrency address format
func ValidateAddress(address string, currency string) bool {
	switch currency {
	case "BTC":
		// Bitcoin testnet addresses
		return len(address) >= 26 && (address[:2] == "tb" || address[0] == 'n' || address[0] == 'm')
	case "ETH", "USDT":
		// Ethereum addresses
		return len(address) == 42 && address[:2] == "0x"
	default:
		return len(address) > 0
	}
}
