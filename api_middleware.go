package main

import (
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

func createTokenAuthMiddleware(authToken string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawToken := r.Header.Get("Authorization")
			if rawToken != "" {
				tokenParts := strings.Split(rawToken, " ")
				if len(tokenParts) == 2 {
					if tokenParts[0] == "Bearer" {
						reqToken := tokenParts[1]
						if reqToken == authToken {
							next.ServeHTTP(w, r)
							return
						}
					}
				}
			}

			// Write an error and stop the handler chain
			http.Error(w, "Forbidden", http.StatusForbidden)
		})
	}

}
