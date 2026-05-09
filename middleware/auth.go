package middleware

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func init() {
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		secret = "change-me-in-production-use-SESSION_SECRET-env"
	}
	store = sessions.NewCookieStore([]byte(secret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400,
		HttpOnly: true,
		Secure:   os.Getenv("APP_ENV") == "production",
	}
}

func SessionAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "b2b-session")
		if err != nil {
			writeUnauthorized(w)
			return
		}

		authenticated, ok := session.Values["authenticated"].(bool)
		if !ok || !authenticated {
			writeUnauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
}
