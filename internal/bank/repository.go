package bank

import (
	"database/sql"
	"errors"
	"plugin/internal/database/queries"
)

type Repository interface {
	Create(initialBalance int) error
	GetBalance() (int, error)
	SetBalance(amount int) error
	Deposit(amount int) error
	Withdraw(amount int) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(initialBalance int) error {
	_, err := r.db.Exec(queries.CreateBank, initialBalance)
	return err
}

func (r *repository) GetBalance() (int, error) {
	var balance int
	err := r.db.QueryRow(queries.GetBankBalance).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}

func (r *repository) SetBalance(amount int) error {
	_, err := r.db.Exec(queries.SetBankBalance, amount)
	return err
}

func (r *repository) Deposit(amount int) error {
	_, err := r.db.Exec(queries.DepositToBank, amount)
	return err
}

func (r *repository) Withdraw(amount int) error {
	_, err := r.db.Exec(queries.WithdrawFromBank, amount)
	return err
}
