package midlware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"library/internal/auth"
)

func AuthMiddleware(requiredRole string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			authHeader := r.Header.Get("Authorization")
			tokenString := strings.TrimPrefix(authHeader,"Bearer ")
			
			claims := &auth.MyClaims{}
			token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(auth.SecretKey), nil
			})
			if err !=nil || !token.Valid{
				http.Error(w,"Невалидный токен", http.StatusUnauthorized)
				return
			}

			if requiredRole != "" {
				 if claims.Role != "admin" && claims.Role != requiredRole {
					http.Error(w, "Доступ запрещён", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w,r)
		})

	}
}

