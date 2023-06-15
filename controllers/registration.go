package controllers

import (
	"fmt"
	"lms-application/config"
	"lms-application/middleware"
	"lms-application/models"
	"log"
	"net/http"
	"os"
	"strconv"
	
	"github.com/go-gomail/gomail"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
)

func RegistrationStepOne(c echo.Context) error {
	var user models.User
	var user_profile models.UserProfile

	if errBind := c.Bind(&user); errBind != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "error": errBind.Error(),
        })
	}
	
	if errSave := config.DB.Save(&user).Error; errSave != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "failed create user!")
	}else{
		input := map[string]interface{}{
			"key_name": "registration_step",
			"value" : 1,
		}

		config.DB.Model(&user_profile).Create(&input)
	}

	token, errToken := middleware.CreateToken(int(user.ID))

	if errToken != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "registration create jwt failed",
			"error":   errToken.Error(),
		})
	}

	errLoad := godotenv.Load(".env")

	if errLoad != nil {
		log.Fatalf("Error loading .env file")
	}

	publicKeyPEM := []byte(os.Getenv("PUBLIC_KEY_JWT"))

	publicKey, errParse := jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	
	if errParse != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to parse public key")
	}

	tokenParse, err := jwt.Parse(token, func(tokenParse *jwt.Token) (interface{}, error) {
		if _, ok := tokenParse.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", tokenParse.Header["alg"])
		}
		return publicKey, nil
	})

	if claims, ok := tokenParse.Claims.(jwt.MapClaims); ok && tokenParse.Valid {
		userId := claims["userId"]

		if err := config.DB.Where("id = ?", userId).First(&user).Error; err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Record not found!")
		}
	}

	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("MAIL_FROM"))
	m.SetHeader("To", user.Email)
	m.SetHeader("Subject", "Email Verification LMS Application")

	body := "<h1>This is the email body.</h1><p>This email contains HTML content.</p>"

	m.SetBody("text/html", body)

	mailPort, err := strconv.Atoi(os.Getenv("MAIL_PORT"))

	if err != nil {
		return err
	}

	d := gomail.NewDialer(os.Getenv("MAIL_HOST"), mailPort, os.Getenv("MAIL_USERNAME"), os.Getenv("MAIL_PASSWORD"))

	errDial := d.DialAndSend(m)
	if errDial != nil {
		return errDial
	}

	userResponse := models.UserResponse{user, token}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "success create new user",
		"user":   userResponse,
	})
}