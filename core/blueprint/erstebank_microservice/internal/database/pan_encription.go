package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// Encrypt encrypts plain text using AES-256-GCM
func Encrypt(plainText string) (string, error) {
	// 1. Čitanje ključa iz ENV
	key := []byte(os.Getenv("BANK_ENCRYPTION_KEY"))
	
	// 2. Kreiranje cipher bloka
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Error creating cipher block:", err)
		return "", err
	}

	// 3. GCM mod rada (preporučen za PCI DSS)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		fmt.Println("Error creating GCM:", err)
		return "", err
	}

	// 4. Kreiranje Nonce-a (jedinstveni broj za svaku enkripciju)
	// PCI DSS zahteva da se isti ključ ne koristi isto bez nasumičnog dela
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println("Error generating nonce:", err)
		return "", err
	}

	// 5. Enkripcija (Sealing)
	// Spajamo nonce i šifrovani tekst radi lakšeg čuvanja u bazi
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	return hex.EncodeToString(ciphertext), nil
}

// Decrypt decrypts hex string using AES-256-GCM
func Decrypt(encryptedHex string) (string, error) {
	key := []byte(os.Getenv("BANK_ENCRYPTION_KEY"))
	
	ciphertext, _ := hex.DecodeString(encryptedHex)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Odvajanje nonce-a od podataka
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	
	plainText, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}