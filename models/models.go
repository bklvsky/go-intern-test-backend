package models

import (
	"fmt"
	"time"
)

var ErrUserNotFound = fmt.Errorf("User not found")
var ErrNotEnoughCredit = fmt.Errorf("Not enough money in the account")

type User struct {
	ID      int     `json:"id"`
	Balance float32 `json:"balance"`
	Reserve float32 `json:"-"`
}

type Users []*User

type Transaction struct {
	ID           int       `json:"-"`
	OrderId      int       `json:"orderId"`
	UserId       int       `json:"clientId"`
	ServiceId    int       `json:"serviceId"`
	Value        float32   `json:"value"`
	ReserveValue float32   `json:"-"`
	Timesp       time.Time `json:"time"`
	Status       string    `json:"status,omitempty"`
	Note         string    `json:"note,omitempty"`
}

type Transactions []*Transaction

type Transfer struct {
	Sender    int     `json:"sender"`
	Recipient int     `json:"recipient"`
	Value     float32 `json:"value"`
}

type History struct {
	History Transactions `json:"history"`
}

type HistoryRequest struct {
	UserId int    `json:"userId"`
	Page   int    `json:"page"`
	Sort   string `json:"sort"`
}
