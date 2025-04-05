package auth

import "net/http"

func noAuth(next http.Handler) http.Handler {
	return next
}
