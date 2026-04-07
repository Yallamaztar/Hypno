package wallet

import (
	"database/sql"
	"errors"
	"plugin/internal/database/queries"
	"sync"
)

// wallet represents a player's wallet
type wallet struct {
	PlayerID int
	Balance  int
	Name     string
}

type repository struct {
	db *sql.DB
	mu sync.Mutex
}

type Repository interface {
	Create(playerID int, startingBalance int) error
	Balance(playerID int) (int, error)
	SetBalance(playerID int, amount int) error
	Deposit(playerID int, amount int) error
	Withdraw(playerID int, amount int) error
	Delete(playerID int) error
	Exists(playerID int) (bool, error)
	GetTopWallets(limit int) ([]wallet, error)
	GetBottomWallets(limit int) ([]wallet, error)
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

// Create creates a new wallet for a player
func (r *repository) Create(playerID int, startingBalance int) error {
	res, err := r.db.Exec(queries.CreateWallet, playerID, startingBalance)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()
	return err
}

// Balance retrieves the balance of a players wallet
func (r *repository) Balance(playerID int) (int, error) {
	var balance int
	err := r.db.QueryRow(queries.GetWalletBalanceByPlayerID, playerID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}
	return balance, nil
}

// SetBalance sets the balance of a players wallet
func (r *repository) SetBalance(playerID int, amount int) error {
	_, err := r.db.Exec(queries.SetWalletBalance, amount, playerID)
	return err
}

// Deposit deposits money into a players wallet
func (r *repository) Deposit(playerID int, amount int) error {
	_, err := r.db.Exec(queries.DepositToWallet, amount, playerID)
	return err
}

func (r *repository) Withdraw(playerID int, amount int) error {
	_, err := r.db.Exec(queries.WithdrawFromWallet, amount, playerID)
	return err
}

func (r *repository) Delete(playerID int) error {
	_, err := r.db.Exec(queries.DeleteWallet, playerID)
	return err
}

func (r *repository) Exists(playerID int) (bool, error) {
	var count int
	err := r.db.QueryRow(queries.WalletExists, playerID).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *repository) GetTopWallets(limit int) ([]wallet, error) {
	return r.getWallets(queries.GetTopWallets, limit)
}

func (r *repository) GetBottomWallets(limit int) ([]wallet, error) {
	return r.getWallets(queries.GetBottomWallets, limit)
}

func (r *repository) getWallets(query string, limit int) ([]wallet, error) {
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []wallet
	for rows.Next() {
		var w wallet
		if err := rows.Scan(&w.PlayerID, &w.Balance, &w.Name); err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return wallets, nil
}
