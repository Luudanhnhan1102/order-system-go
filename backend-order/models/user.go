package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID            primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email         string             `json:"email" bson:"email"`
	Password      string             `json:"-" bson:"password"` // The "-" tag means this field won't be included in JSON output
	IsAdmin       bool               `json:"isAdmin" bson:"isAdmin"`
	ResetToken    string             `json:"-" bson:"resetToken,omitempty"`
	ResetTokenExp time.Time          `json:"-" bson:"resetTokenExp,omitempty"`
}

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
