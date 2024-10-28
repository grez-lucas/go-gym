package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

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

		token, err := ValidateJWT(tokenString)

		if err != nil {
			log.Printf("Error validating JWT: `%v`", err.Error())
			WriteUnauthorized(w)
			return
		}

		// Check if token is valid to extract claims
		if !token.Valid {
			log.Printf("Token is invalid \n")
			WriteUnauthorized(w)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)

		if !ok {
			log.Printf("Token has invalid claims \n")
			WriteUnauthorized(w)
			return
		}

		accountID := int64(claims["accountID"].(float64))

		// Store the ID in GoLang context
		// So that we can pass it around to later methods which require auth

		ctx := context.WithValue(req.Context(), accountID, accountID)

		handlerFunc(w, req.WithContext(ctx))
	}
}

func CreateJWT(account *Account) (string, error) {

	claims := &jwt.MapClaims{
		"expiresAt": 15000,
		"accountID": account.ID,
	}

	secret := LoadConfig().JWTSecret
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func ValidateJWT(tokenString string) (*jwt.Token, error) {

	jwtSecret := LoadConfig().JWTSecret

	return jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: `%v`", t.Header["&alg"])
		}
		return []byte(jwtSecret), nil
	})

}
