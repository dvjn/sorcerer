package auth

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/dvjn/sorcerer/internal/config"
	"github.com/go-chi/chi/v5"
)

func proxyHeaderAuthMiddleware(c *config.ProxyHeaderAuth) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := r.Header.Get(c.UserHeaderName)
			groups := r.Header.Get(c.GroupsHeaderName)
			groupsList := strings.Split(groups, c.GroupsHeaderSeparator)

			owner := chi.URLParam(r, "owner")

			if owner == user || slices.Contains(groupsList, owner) {
				fmt.Printf("User %s (%s) is authorized to access the resource %s\n", user, groups, owner)
				next.ServeHTTP(w, r)
				return
			}
			fmt.Printf("User %s (%s) is not authorized to access the resource %s\n", user, groups, owner)

			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		})
	}
}
