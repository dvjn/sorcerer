package proxy_header_auth

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/dvjn/sorcerer/internal/config"
	"github.com/go-chi/chi/v5"
)

type ProxyHeaderAuth struct {
	c *config.ProxyHeaderAuthConfig
}

func New(c *config.ProxyHeaderAuthConfig) *ProxyHeaderAuth {
	return &ProxyHeaderAuth{c}
}

func (a *ProxyHeaderAuth) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get(a.c.UserHeaderName)
		groups := r.Header.Get(a.c.GroupsHeaderName)
		groupsList := strings.Split(groups, a.c.GroupsHeaderSeparator)

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
