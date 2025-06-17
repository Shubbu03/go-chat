package pkg

import (
	"fmt"

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
