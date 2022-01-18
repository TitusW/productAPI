package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/TitusW/productAPI/helpers"
)

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientToken := r.Header.Get("token")
		if clientToken == "" {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode("You do not have the authorization to this page")
			return
		}

		claims, err := helpers.ValidateToken(clientToken)
		if err != "" {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode("Failed to validate")
			return
		}
		w.Header().Add("Email", claims.Email)
		w.Header().Add("First_name", claims.First_name)
		w.Header().Add("Last_name", claims.Last_name)
		w.Header().Add("user_type", "ADMIN")
		w.Header().Add("Uid", claims.Uid)
		next.ServeHTTP(w, r)
	})
}
