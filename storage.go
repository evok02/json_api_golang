package main

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
)

type Storage interface {
	CreateAccount(*Account) error
	DeleteAccount(int) error
	UpdateAccount(*Account) error
	GetAccountByID(int) (*Account, error)
	GetAccounts() ([]*Account, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore() (*PostgresStore, error) {
	connStr := "user=postgres dbname=postgres password=gobank sslmode=disable"
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		return nil, fmt.Errorf("couldnt open the connection: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("couldnt open the connection: %w", err)
	}

	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) Init() error {
	return s.createAccountTable()
}

func (s *PostgresStore) createAccountTable() error {
	q := `
	CREATE TABLE IF NOT EXISTS account (
	id SERIAL PRIMARY KEY, 
	first_name varchar(55) NOT NULL,
	last_name varchar(55) NOT NULL,
	bank_number int NOT NULL DEFAULT 0,
	balance int DEFAULT 0,
	created_at timestamp DEFAULT NOW() NOT NULL
	)
	`
	_, err := s.db.Exec(q)
	if err != nil {
		return fmt.Errorf("couldn't create account table: %w", err)
	}

	return nil
}

func (s *PostgresStore) CreateAccount(a *Account) error {
	q1 := `
	
	INSERT INTO account (first_name, last_name, bank_number, balance, created_at)
	VALUES ($1, $2, $3, $4, $5);
	`
	res1, err := s.db.Exec(q1, a.FirstName, a.LastName, a.BankNumber, a.Balance, a.CreatedAt)
	if err != nil {
		return fmt.Errorf("couldnt execute insert query: %w", err)
	}

	raws, err := res1.RowsAffected()
	if err != nil {
		return fmt.Errorf("couldnt fetch the result from the storage: %w", err)
	}

	if raws == 0 {
		return errors.New("didn't affect any rows, query was usuccesful")
	}

	q2 := `COMMIT;`
	_, err = s.db.Exec(q2)
	if err != nil {
		return fmt.Errorf("couldnt execute insert query: %w", err)
	}

	return nil
}

func (s *PostgresStore) UpdateAccount(a *Account) error {
	return nil
}

func (s *PostgresStore) GetAccounts() ([]*Account, error) {
	q := `
	SELECT * FROM account
	`

	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("couldnt execute insert query: %w", err)
	}

	accounts := []*Account{}
	for rows.Next() {
		account, err := rowsIntoAccount(rows)
		if err != nil {
			return nil, fmt.Errorf("couldn't scan the row: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func rowsIntoAccount(r Scanner) (*Account, error) {
	account := &Account{}
	err := r.Scan(
		&account.ID,
		&account.FirstName,
		&account.LastName,
		&account.BankNumber,
		&account.Balance,
		&account.CreatedAt)

	return account, err
}

func (s *PostgresStore) DeleteAccount(id int) error {
	q := `
	DELETE FROM account
	WHERE id = $1
	`

	res, err := s.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("couldnt execute the query: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("couldnt check the result of the query: %w", err)
	}

	if n == 0 {
		return fmt.Errorf("0 rows affected")
	}

	return nil
}

func (s *PostgresStore) GetAccountByID(id int) (*Account, error) {
	q := `
	SELECT * FROM account
	WHERE id = $1
	`
	row := s.db.QueryRow(q, id)

	account := &Account{}
	account, err := rowsIntoAccount(row)

	if err != nil {
		return nil, fmt.Errorf("couldn't scan the row: %w", err)
	}

	return account, nil
}
