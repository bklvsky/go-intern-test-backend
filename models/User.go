package models

import (
	// "log"
	"fmt"
	"io"
)

type User struct {
	ID      int     `json:"id"`
	Balance float32 `json:"balance"`
	Reserve float32 `json:"reserve"`
}

type Users []*User

var ErrUserNotFound = fmt.Errorf("User not found")

type UserRepository interface {
	ToJSON(wr io.Writer) error
	FindUserByID(id int) (*User, error)
	FindAllUsers() ([]*User, error)
	UpdateUserBalance(newValue *User) error
	AddUser(newValue *User) error
	// PostUserBalance(id int, balance float32)
}
