package pkg

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func HashPassword(passwordToHash string) []byte {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordToHash), bcrypt.DefaultCost)

	if err != nil {
		fmt.Printf("Error occured while hashing password! %s", err)
	}

	return hashedPassword
}

func ComparePassword(passwordToCompare string, dbPassword []byte) bool {
	err := bcrypt.CompareHashAndPassword([]byte(passwordToCompare), dbPassword)

	if err != nil {
		fmt.Printf("Passwords not same! %s", err)
		return false
	}

	return true
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"error": message,
	})
}

func ToString(v interface{}) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", v)
}
