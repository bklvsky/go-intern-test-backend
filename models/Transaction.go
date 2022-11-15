package models

import "time"

type Transaction struct {
	ID int `json:"-"`
	OrderId int `json:"orderId"`
	UserId int `json:"clientId"`
	ServiceId int `json:"serviceId"`
	Value float32 `json:"value"`
	ReserveValue float32 `json:"-"`
	Timesp time.Time `json:"time"`
	Note string `json:"note"`
}



type Transactions []*Transaction

type Transfer struct {
	Sender int `json:"sender"`
	Recipient int `json:"recipient"`
	Value float32 `json:"value"`
}