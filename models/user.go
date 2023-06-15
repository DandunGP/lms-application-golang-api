package models

import (
	"gorm.io/gorm"
)

type EmailStatus string

const (
    EmailStatusWaiting   EmailStatus = "waiting"
    EmailStatusVerified  EmailStatus = "verified"
)

type User struct {
	gorm.Model
	Email string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
	Gender string `json:"gender" form:"gender"`
	Name string `json:"name" form:"name"`
	ProfilePicture string `json:"profile_picture" form:"profile_picture"`
	ActivationCode string `json:"activation_code" form:"activation_code"`
	EmailStatus string
	SocialID int
}

type UserResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}