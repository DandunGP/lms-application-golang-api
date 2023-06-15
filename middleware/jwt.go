package middleware

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

func CreateToken(userId int) (string, error) {
	errLoad := godotenv.Load(".env")

	if errLoad != nil {
		log.Fatalf("Error loading .env file")
	}
	
	claims := jwt.MapClaims{}
	claims["userId"] = userId
	claims["exp"] = time.Now().Add(time.Hour * 1).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("PRIVATE_KEY_JWT")))
}