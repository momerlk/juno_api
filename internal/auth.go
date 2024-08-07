package internal

import (
	"time"

	"net/http"
	
	jwt "github.com/golang-jwt/jwt/v5"

)

func GenerateToken(UserId string) (string, error){
	secret := Getenv("JWT_KEY")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":    UserId,
			"session_id": GenerateId(),
			"exp":        time.Now().Add(20 * time.Minute).Unix(),
		})

	tokenString, err := token.SignedString([]byte(secret))
	return tokenString , err;
}

func GenerateRefreshToken(UserId string) (string, error){
	secret := Getenv("JWT_KEY")
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":    UserId,
			"session_id": GenerateId(),
			"exp":        time.Now().Add((7*24) * time.Hour).Unix(), // valid till 7 days
		})

	tokenString, err := token.SignedString([]byte(secret))
	return tokenString , err;
}


func ParseTokenString(tokenString string) (jwt.MapClaims , bool){
	var tokenClaims jwt.MapClaims

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		key := Getenv("JWT_KEY")
		return []byte(key) , nil
	})
	if err != nil {
		return nil , false
	}

	// Check if the token is valid and not expired
	if claims , ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return nil , false
	} else {
		tokenClaims = claims
	}

	return tokenClaims , true;
}

func Verify(w http.ResponseWriter , r *http.Request) (jwt.MapClaims , bool) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Authorization token missing", http.StatusUnauthorized)
		return nil , false
	}

	var tokenClaims jwt.MapClaims

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		key := Getenv("JWT_KEY")
		return []byte(key) , nil
	})
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return nil , false
	}

	// Check if the token is valid and not expired
	if claims , ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		http.Error(w, "Invalid token (expired)", http.StatusUnauthorized)
		return nil , false
	} else {
		tokenClaims = claims
	}


	return tokenClaims , true
}

