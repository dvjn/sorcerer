package auth

import "net/http"

func noAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return next
	}
}
