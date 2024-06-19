package internal 

import (
	"fmt"
	"os"

	"crypto/sha256"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

func Getenv(key string) string{
	return os.Getenv(key)
}

func Hash(s string) string {
	return fmt.Sprintf("%x" , sha256.Sum256([]byte(s)))
}

func HashAndSalt(b []byte) (string , error) {
	hash , err := bcrypt.GenerateFromPassword(b , bcrypt.MinCost)
	if err != nil {
		return "" , err
	}
	
	return string(hash) , nil
}


func GenerateId() string {
	return uuid.NewString()
}