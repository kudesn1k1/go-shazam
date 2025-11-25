package user

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"go-shazam/internal/auth"
	"io"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrDecryptionFailed  = errors.New("decryption failed")
)

// CryptoService handles email encryption/decryption and password hashing
type CryptoService struct {
	emailEncryptionKey []byte // 32 bytes for AES-256
}

func NewCryptoService(config *auth.Config) *CryptoService {
	// AES-256 requires exactly 32 bytes key
	// Hash the key to ensure it's always 32 bytes
	key := sha256.Sum256([]byte(config.EmailEncryptionKey))

	return &CryptoService{
		emailEncryptionKey: key[:],
	}
}

// HashEmail creates a deterministic SHA-256 hash of the email for lookups
// Email is normalized (lowercase, trimmed) before hashing
func (c *CryptoService) HashEmail(email string) string {
	normalized := strings.ToLower(strings.TrimSpace(email))
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:])
}

// EncryptEmail encrypts the email using AES-256-GCM
func (c *CryptoService) EncryptEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))

	block, err := aes.NewCipher(c.emailEncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(normalized), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptEmail decrypts the email using AES-256-GCM
func (c *CryptoService) DecryptEmail(encryptedEmail string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedEmail)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(c.emailEncryptionKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrDecryptionFailed
	}

	return string(plaintext), nil
}

// HashPassword creates a bcrypt hash of the password
func (c *CryptoService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// VerifyPassword checks if the password matches the hash
func (c *CryptoService) VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
