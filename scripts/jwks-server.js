#!/usr/bin/env node
/**
 * Simple JWKS HTTP server for testing.
 * Serves a pre-generated JWKS JSON at /.well-known/jwks.json
 */

const http = require('http');
const fs = require('fs');
const path = require('path');

const PORT = process.env.PORT || 8888;
const JWKS_FILE = process.env.JWKS_FILE || 'test-keys/jwks.json';

const server = http.createServer((req, res) => {
  console.log(`${new Date().toISOString()} - ${req.method} ${req.url}`);

  if (req.url === '/.well-known/jwks.json' || req.url === '/jwks.json') {
    try {
      const jwks = JSON.parse(fs.readFileSync(JWKS_FILE, 'utf8'));
      res.writeHead(200, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify(jwks));
    } catch (err) {
      res.writeHead(500, { 'Content-Type': 'application/json' });
      res.end(JSON.stringify({ error: 'Failed to read JWKS' }));
    }
  } else if (req.url === '/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ status: 'ok' }));
  } else {
    res.writeHead(404, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ error: 'Not found' }));
  }
});

server.listen(PORT, () => {
  console.log(`JWKS server listening on port ${PORT}`);
  console.log(`Serving JWKS from: ${JWKS_FILE}`);
  console.log(`Endpoints:`);
  console.log(`  http://localhost:${PORT}/.well-known/jwks.json`);
  console.log(`  http://localhost:${PORT}/jwks.json`);
});

// Graceful shutdown
process.on('SIGTERM', () => {
  console.log('Received SIGTERM, shutting down...');
  server.close(() => {
    console.log('Server closed');
    process.exit(0);
  });
});
