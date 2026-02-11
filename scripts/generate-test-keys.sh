#!/bin/bash
set -e

# Generate RSA key pair for testing
KEY_SIZE=2048
KEY_DIR="./test-keys"

echo "Generating RSA key pair for JWKS testing..."

mkdir -p "$KEY_DIR"

# Generate private key in PEM format
openssl genrsa -out "$KEY_DIR/private.pem" "$KEY_SIZE"

# Convert to PKCS8 format (required by some JWT libraries)
openssl pkcs8 -topk8 -inform PEM -outform PEM -nocrypt -in "$KEY_DIR/private.pem" -out "$KEY_DIR/private-pkcs8.pem"

# Generate public key in PEM format
openssl rsa -pubout -in "$KEY_DIR/private.pem" -out "$KEY_DIR/public.pem"

# Generate JWK from public key
# Using jwk-cli or Node.js with jose library

echo "Key generation complete."
echo "Private key: $KEY_DIR/private.pem"
echo "Public key: $KEY_DIR/public.pem"
