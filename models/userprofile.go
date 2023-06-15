package models

import "gorm.io/gorm"

type UserProfile struct {
	gorm.Model
	UserID uint `json:"user_id" form:"user_id"`
	User User
	KeyName string `json:"key_name" form:"key_name"`
	Value string `json:"value" form:"value"`
}