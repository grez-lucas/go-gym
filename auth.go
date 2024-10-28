package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

func WriteUnauthorized(w http.ResponseWriter) {

	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(http.StatusText(http.StatusUnauthorized)))

}

// To decorate certain HTTP handlers with JWT authentication (the ones who
// require it)

func WithJWTAuth(handlerFunc http.HandlerFunc) http.HandlerFunc {

	log.Println("Calling JWT middleware")

	return func(w http.ResponseWriter, req *http.Request) {

		tokenString := req.Header.Get("x-jwt-token")

		if _, err := ValidateJWT(tokenString); err != nil {
			WriteJSON(w, http.StatusForbidden, APIError{"Invalid token"})
			return
		}

		handlerFunc(w, req)
	}
}

func CreateJWT(account *Account) (string, error) {

	claims := &jwt.MapClaims{
		"expiresAt":   15000,
		"accountName": account.UserName,
	}

	secret := LoadConfig().JWTSecret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {

	jwtSecret, found := os.LookupEnv("JWT_SECRET")

	if !found {
		return nil, fmt.Errorf("Couldn't find environment variable JWT_SECRET")
	}

	return jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: `%v`", t.Header["&alg"])
		}
		return []byte(jwtSecret), nil
	})

}
