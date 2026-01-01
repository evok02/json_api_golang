package main

import (
	"math/rand"
	"time"
)

type Transfer struct {
	ToAccount int `json:"to_account"`
	Amount    int `json:"amount"`
}

type Account struct {
	ID         int64     `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	BankNumber int64     `json:"bank_number"`
	Balance    int64     `json:"balance"`
	CreatedAt  time.Time `json:"created_at"`
}

type ReqCreateAccount struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type Scanner interface {
	Scan(dest ...any) error
}

func NewAccount(firstName, lastName string) *Account {
	return &Account{
		FirstName:  firstName,
		LastName:   lastName,
		BankNumber: int64(rand.Intn(100000000)),
		CreatedAt:  time.Now().UTC(),
	}
}
