package models

import (
	// "log"
	"fmt"
)

type User struct {
	ID      int     `json:"id"`
	Balance float32 `json:"balance"`
	Reserve float32 `json:"reserve"`
}

type Users []*User

var ErrUserNotFound = fmt.Errorf("User not found")
var ErrNotEnoughCredit = fmt.Errorf("Not enough money in the account")
