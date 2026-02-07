package htpasswd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dvjn/sorcerer/internal/config"
	"github.com/rs/zerolog"
)

// Valid bcrypt hash for "password" generated with htpasswd -B
const (
	testPassword   = "password"
	testBcryptHash = "$2b$12$1PqeG8v5YfoxsyW5gAyHcOq6RCgY71kIt6qtLnEUqaddiuNGTGepe"
)

func TestHtpasswdAuthNew(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())

	// Test inline content
	cfg := &config.HtpasswdConfig{
		Contents: "testuser:" + testBcryptHash,
	}

	auth, err := NewHtpasswdAuth(cfg, &logger)
	if err != nil {
		t.Fatalf("Failed to create htpasswd auth: %v", err)
	}
	if auth == nil {
		t.Fatal("Expected non-nil auth")
	}
}

func TestHtpasswdAuthMatch(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())

	cfg := &config.HtpasswdConfig{
		Contents: "testuser:" + testBcryptHash + "\notheruser:" + testBcryptHash,
	}

	auth, err := NewHtpasswdAuth(cfg, &logger)
	if err != nil {
		t.Fatalf("Failed to create htpasswd auth: %v", err)
	}

	// Test correct password
	if !auth.Match("testuser", testPassword) {
		t.Error("Expected valid credentials to match")
	}

	// Test wrong password
	if auth.Match("testuser", "wrongpassword") {
		t.Error("Expected wrong credentials to not match")
	}

	// Test non-existent user
	if auth.Match("nonexistent", testPassword) {
		t.Error("Expected non-existent user to not match")
	}
}

func TestHtpasswdAuthConfigValidation(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())

	// Test missing both file and contents
	cfg := &config.HtpasswdConfig{}
	_, err := NewHtpasswdAuth(cfg, &logger)
	if err == nil {
		t.Error("Should fail when neither file nor contents provided")
	}
	if !strings.Contains(err.Error(), "neither file nor contents provided") {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestHtpasswdAuthMultipleUsers(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())

	// Multiple users with different hash types
	cfg := &config.HtpasswdConfig{
		Contents: "user1:" + testBcryptHash + "\n" +
			"user2:$apr1$r31so$NwH8HLi4U/bD5E0E2L4bK/\n" +
			"user3:$5$salt$hash\n" +
			"# This is a comment\n",
	}

	auth, err := NewHtpasswdAuth(cfg, &logger)
	if err != nil {
		t.Fatalf("Failed to create htpasswd auth: %v", err)
	}

	// Test each user
	if !auth.Match("user1", testPassword) {
		t.Error("user1 credentials should match")
	}

	// These won't match because the hash value doesn't match "password"
	if auth.Match("user2", "password") {
		t.Error("user2 with invalid hash should not match")
	}
}

func TestBasicAuthMiddleware(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	cfg := &config.HtpasswdConfig{
		Contents: "testuser:" + testBcryptHash,
	}

	auth, err := NewHtpasswdAuth(cfg, &logger)
	if err != nil {
		t.Fatalf("Failed to create auth: %v", err)
	}

	middleware := auth.DistributionMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if username, ok := GetUsernameFromContext(r.Context()); ok {
			w.Write([]byte("Hello, " + username))
		} else {
			w.Write([]byte("Hello, stranger"))
		}
	}))

	// Test no credentials
	req := httptest.NewRequest("GET", "/v2/some/path", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test wrong credentials
	req.SetBasicAuth("testuser", "wrongpassword")
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	// Test correct credentials
	req.SetBasicAuth("testuser", testPassword)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "Hello, testuser") {
		t.Errorf("Expected greeting message, got: %s", body)
	}
}

func TestShouldSkipAuth(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	cfg := &config.HtpasswdConfig{
		Contents: "testuser:" + testBcryptHash,
	}

	auth, err := NewHtpasswdAuth(cfg, &logger)
	if err != nil {
		t.Fatalf("Failed to create auth: %v", err)
	}

	middleware := auth.DistributionMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))

	// Skip paths to test
	skipPaths := []string{"/v2/", "/health", "/metrics"}

	for _, path := range skipPaths {
		req := httptest.NewRequest("GET", path, nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("Path %s should skip auth, got status %d", path, w.Code)
		}
	}

	// Path that requires auth
	req := httptest.NewRequest("GET", "/v2/some/repo/tags/list", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Path /v2/some/repo/tags/list should require auth, got status %d", w.Code)
	}
}

func TestInvalidCredentialsAgainstAuthSpec(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	cfg := &config.HtpasswdConfig{
		Contents: "testuser:" + testBcryptHash,
	}

	auth, err := NewHtpasswdAuth(cfg, &logger)
	if err != nil {
		t.Fatalf("Failed to create auth: %v", err)
	}

	middleware := auth.DistributionMiddleware()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	testCases := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
	}{
		{
			name:           "wrong password for existing user",
			username:       "testuser",
			password:       "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "non-existent user with password",
			username:       "nonexistent",
			password:       "password",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "non-existent user empty password",
			username:       "nonexistent",
			password:       "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "existing user empty password",
			username:       "testuser",
			password:       "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "non-existent user with very long password",
			username:       "hacker",
			password:       strings.Repeat("a", 1000),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v2/some/repo/manifests/latest", nil)
			req.SetBasicAuth(tc.username, tc.password)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			// Verify 401 status
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			// Verify WWW-Authenticate header is set (per auth spec)
			authHeader := w.Header().Get("WWW-Authenticate")
			if authHeader == "" {
				t.Error("Expected WWW-Authenticate header to be set")
			}
			if !strings.Contains(authHeader, "Sorcerer OCI Registry") {
				t.Errorf("Expected WWW-Authenticate to contain realm, got: %s", authHeader)
			}

			// Verify that response body doesn't reveal user existence
			// (should be the same for wrong password as non-existent user)
			body1 := w.Body.String()

			// Test with a different invalid credential
			req2 := httptest.NewRequest("GET", "/v2/some/repo/manifests/latest", nil)
			req2.SetBasicAuth("otheruser", "otherpass")
			w2 := httptest.NewRecorder()
			handler.ServeHTTP(w2, req2)

			body2 := w2.Body.String()
			if body1 != body2 {
				t.Error("Response should not reveal whether user exists by differing")
			}
		})
	}
}

func TestGetUsernameFromContext(t *testing.T) {
	// Test with no context
	ctx := context.Background()
	username, ok := GetUsernameFromContext(ctx)
	if ok {
		t.Error("Should not find username in default context")
	}
	if username != "" {
		t.Errorf("Expected empty username, got %s", username)
	}

	// Test with context
	ctx = context.WithValue(context.Background(), userContextKey, "testuser")
	username, ok = GetUsernameFromContext(ctx)
	if !ok {
		t.Error("Should find username in context")
	}
	if username != "testuser" {
		t.Errorf("Expected username %s, got %s", "testuser", username)
	}
}

func TestHtpasswdAuthEmptyContent(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())

	// Empty content should fail
	cfg := &config.HtpasswdConfig{
		Contents: "",
	}

	_, err := NewHtpasswdAuth(cfg, &logger)
	if err == nil {
		t.Error("Should fail with empty content")
	}
}

func TestHtpasswdAuthInvalidContent(t *testing.T) {
	logger := zerolog.New(zerolog.NewConsoleWriter())

	// Invalid htpasswd format
	cfg := &config.HtpasswdConfig{
		Contents: "invalid content without colon",
	}

	auth, err := NewHtpasswdAuth(cfg, &logger)
	// The library might accept this or reject it depending on implementation
	// We just check that it doesn't panic
	if err != nil {
		// Expected to fail with invalid format
		return
	}
	if auth == nil {
		t.Fatal("Expected non-nil auth even with invalid content")
	}
}
