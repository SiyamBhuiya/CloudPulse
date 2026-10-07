package middleware

import (
	"context"
	"net/http"
	"strings"
)

type ctxKey struct{}

// Auth rejects requests without a valid "Authorization: Bearer <token>" header.
func Auth(parse func(string) (string, error)) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !ok || token == "" {
				unauthorized(w)
				return
			}
			userID, err := parse(token)
			if err != nil {
				unauthorized(w)
				return
			}
			ctx := context.WithValue(r.Context(), ctxKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserID returns the logged-in user's id inside a handler.
func UserID(ctx context.Context) string {
	id, _ := ctx.Value(ctxKey{}).(string)
	return id
}

func unauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
}