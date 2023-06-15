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
	"time"
	"crypto/sha256"
	"encoding/hex"
	
	"github.com/go-gomail/gomail"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"gorm.io/gorm"

)

func RegistrationStepOne(c echo.Context) error {
	var user models.User

	if errBind := c.Bind(&user); errBind != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
            "error": errBind.Error(),
        })
	}

	// Check User
	if errUserProfile := config.DB.Where("email = ?", user.Email).First(&user).Error; errUserProfile != nil {
		if errUserProfile == gorm.ErrRecordNotFound {
				// Generate Activation Code 
				location, err := time.LoadLocation("Asia/Jakarta")
				if err != nil {
					panic(err)
				}
				currentTime := time.Now().In(location)
				dateTimeString := currentTime.Format("20060102150405")
				activationCode := "ACT" + dateTimeString

				// Hash activation code using SHA
				hasher := sha256.New()
				hasher.Write([]byte(activationCode))
				hashedBytes := hasher.Sum(nil)
				hashedString := hex.EncodeToString(hashedBytes)

				user.ActivationCode =  hashedString

				user.EmailStatus = "pending"

				if errSave := config.DB.Save(&user).Error; errSave != nil {
					return echo.NewHTTPError(http.StatusBadRequest, "failed create user!")
				}else{
					userProfile := models.UserProfile{
						UserID : user.ID,
						KeyName : "registration_step",
						Value : "1",
					}

					if errUserProfile := config.DB.Create(&userProfile).Error; errUserProfile != nil {
						return echo.NewHTTPError(http.StatusBadRequest, "failed create user profile!")
					}
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
		return echo.NewHTTPError(http.StatusBadRequest, "failed check user!")
	}

	return c.JSON(http.StatusBadRequest, map[string]interface{}{
		"message": "email already register",
		"user":   user,
	})
}

func RegistrationStepTwo(c echo.Context) error { // Email Verification
	var user models.User
	var userProfile models.UserProfile

	activation_code := c.QueryParam("activation_code")

	// Get User If Activation Code Match
	if err := config.DB.Where("activation_code = ?", activation_code).First(&user).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"message": "activation code doesn't match",
		})
	}else{
		if err := config.DB.Where("user_id = ?", user.ID).Where("key_name", "registration_step").First(&userProfile).Error; err != nil {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"message": "activation code doesn't match",
			})
		}

		userProfile.Value = "2"
		config.DB.Save(&userProfile)

		user.ActivationCode = ""
		user.EmailStatus = "verified"
		config.DB.Save(&user)		

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "email validation success",
		})
	}
}