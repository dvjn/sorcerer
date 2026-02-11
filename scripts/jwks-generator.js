#!/usr/bin/env node
/**
 * Generate a JWKS JSON file from an existing RSA key pair.
 * Usage: node scripts/jwks-generator.js [output-file]
 */

const fs = require('fs');
const path = require('path');

// Input key files
const PRIVATE_KEY_PATH = 'test-keys/private-pkcs8.pem';
const PUBLIC_KEY_PATH = 'test-keys/public.pem';
const KEY_ID = 'test-key-1';
const ISSUER = 'https://test.example.com';
const AUDIENCE = 'sorcerer';

// Output file
const OUTPUT_FILE = process.argv[2] || 'test-keys/jwks.json';
const TOKEN_OUTPUT_FILE = 'test-keys/test-token.txt';

async function main() {
  const jose = require('jose');

  console.log('Generating JWKS from RSA keys...');

  // Read the private key
  const privateKeyPem = fs.readFileSync(PRIVATE_KEY_PATH, 'utf8');
  const privateKey = await jose.importPKCS8(privateKeyPem);

  // Read the public key
  const publicKeyPem = fs.readFileSync(PUBLIC_KEY_PATH, 'utf8');
  const publicKey = await jose.importSPKI(publicKeyPem);

  // Create JWKS with the public key
  const publicJWK = await jose.exportJWK(publicKey);
  publicJWK.kid = KEY_ID;
  publicJWK.alg = 'RS256';

  const jwks = {
    keys: [publicJWK]
  };

  // Write JWKS file
  fs.writeFileSync(OUTPUT_FILE, JSON.stringify(jwks, null, 2));
  console.log(`JWKS written to: ${OUTPUT_FILE}`);

  // Generate a test JWT token
  const now = Math.floor(Date.now() / 1000);
  const payload = {
    sub: 'test-user',
    iss: ISSUER,
    aud: AUDIENCE,
    exp: now + 3600, // 1 hour from now
    iat: now
  };

  const token = await new jose.SignJWT(payload)
    .setProtectedHeader({ alg: 'RS256', kid: KEY_ID })
    .setIssuedAt()
    .setExpirationTime('1h')
    .setIssuer(ISSUER)
    .setAudience(AUDIENCE)
    .setSubject('test-user')
    .sign(privateKey);

  fs.writeFileSync(TOKEN_OUTPUT_FILE, token);
  console.log(`Test token written to: ${TOKEN_OUTPUT_FILE}`);
  console.log(`Bearer: ${token}`);
}

main().catch(err => {
  console.error('Error:', err);
  process.exit(1);
});
