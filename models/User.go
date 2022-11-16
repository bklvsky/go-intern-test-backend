package models

import (
	"fmt"
)

type User struct {
	ID      int     `json:"id"`
	Balance float32 `json:"balance"`
	Reserve float32 `json:"-"`
}

type Users []*User

var ErrUserNotFound = fmt.Errorf("User not found")
var ErrNotEnoughCredit = fmt.Errorf("Not enough money in the account")
