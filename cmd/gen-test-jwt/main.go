package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWK struct {
	Kty string   `json:"kty"`
	Kid string   `json:"kid"`
	Alg string   `json:"alg,omitempty"`
	Use string   `json:"use"`
	N   string   `json:"n"`
	E   string   `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func main() {
	keyID := "test-key-1"
	issuer := "https://test.example.com"
	audience := "sorcerer"
	subject := "test-user"
	keyDir := "test-keys"

	// Create key directory
	if err := os.MkdirAll(keyDir, 0755); err != nil {
		log.Fatalf("Failed to create key directory: %v", err)
	}

	// Generate RSA key pair (2048 bits)
	log.Println("Generating RSA key pair...")
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate RSA key: %v", err)
	}
	publicKey := &privateKey.PublicKey

	// Save private key
	log.Println("Writing private key...")
	privateBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateBytes,
	})
	if err := os.WriteFile(filepath.Join(keyDir, "private.pem"), privatePEM, 0600); err != nil {
		log.Fatalf("Failed to write private key: %v", err)
	}

	// Save public key
	log.Println("Writing public key...")
	publicBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		log.Fatalf("Failed to marshal public key: %v", err)
	}
	publicPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicBytes,
	})
	if err := os.WriteFile(filepath.Join(keyDir, "public.pem"), publicPEM, 0644); err != nil {
		log.Fatalf("Failed to write public key: %v", err)
	}

	// Create JWKS
	log.Println("Creating JWKS...")
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(publicKey.E)).Bytes())

	jwk := JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: keyID,
		Alg: "RS256",
		N:   n,
		E:   e,
	}

	jwks := JWKS{
		Keys: []JWK{jwk},
	}

	jwksBytes, err := json.MarshalIndent(jwks, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal JWKS: %v", err)
	}

	jwksPath := filepath.Join(keyDir, "jwks.json")
	if err := os.WriteFile(jwksPath, jwksBytes, 0644); err != nil {
		log.Fatalf("Failed to write JWKS: %v", err)
	}
	log.Printf("Wrote JWKS to: %s", jwksPath)

	// Generate test JWT token
	log.Println("Generating test JWT token...")
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": subject,
		"iss": issuer,
		"aud": audience,
		"iat": now.Unix(),
		"exp": now.Add(time.Hour).Unix(),
		"nbf": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = keyID

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		log.Fatalf("Failed to sign token: %v", err)
	}

	tokenPath := filepath.Join(keyDir, "token.txt")
	if err := os.WriteFile(tokenPath, []byte(tokenString), 0644); err != nil {
		log.Fatalf("Failed to write token: %v", err)
	}
	log.Printf("Wrote token to: %s", tokenPath)
	log.Printf("Bearer token: %s", tokenString)

	log.Println("\nAll assets generated successfully!")
	log.Println("Files created:")
	log.Printf("  - %s/private.pem", keyDir)
	log.Printf("  - %s/public.pem", keyDir)
	log.Printf("  - %s/jwks.json (serve this at /.well-known/jwks.json)", keyDir)
	log.Printf("  - %s/token.txt", keyDir)
}
