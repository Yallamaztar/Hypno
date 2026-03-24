package stats

import (
	"database/sql"
	"errors"
	"plugin/internal/database/queries"
	"time"
)

/*
 * Player Stats
 */
type PlayerStats struct {
	PlayerID int

	TotalGambles int
	TotalWagered int

	TotalWon  int
	TotalLost int
	Wins      int
	Losses    int

	LastGamble *time.Time
}

type PlayerStatsRepository interface {
	Create(playerID int) error
	Get(playerID int) (*PlayerStats, error)
	RecordWin(playerID int, wager int, won int) error
	RecordLoss(playerID int, wager int, lost int) error
	Reset(playerID int) error
}

type playerStatsRepository struct {
	db *sql.DB
}

func NewPlayerStatsRepository(db *sql.DB) PlayerStatsRepository {
	return &playerStatsRepository{db: db}
}

/*
 * Wallet Stats
 */
type WalletStats struct {
	PlayerID int

	Balance       int
	TotalPaid     int
	TotalReceived int

	DepositCount  int
	WithdrawCount int

	PlayerStats
}

type WalletStatsRepository interface {
	Create(playerID int) error
	Get(playerID int) (*WalletStats, error)

	Deposit(playerID int, amount int) error
	Withdraw(playerID int, amount int) error

	Pay(playerID int, amount int) error
	Receive(playerID int, amount int) error

	Reset(playerID int) error
}

type walletStatsRepository struct {
	db *sql.DB
}

func NewWalletStatsRepository(db *sql.DB) WalletStatsRepository {
	return &walletStatsRepository{db: db}
}

/*
 * Gambling Stats
 */

type GamblingStats struct {
	TotalGambles int
	TotalWagered int
	TotalPaid    int
	LastUpdate   *time.Time
}

type GamblingStatsRepository interface {
	Init() error
	Get() (*GamblingStats, error)
	UpdateAfterGamble(wager int, paid int) error
	Reset() error
}

type gamblingStatsRepository struct {
	db *sql.DB
}

func NewGamblingStatsRepository(db *sql.DB) GamblingStatsRepository {
	return &gamblingStatsRepository{db: db}
}

func (r *playerStatsRepository) Create(playerID int) error {
	_, err := r.db.Exec(queries.CreatePlayerStats, playerID)
	return err
}

func (r *playerStatsRepository) Get(playerID int) (*PlayerStats, error) {
	var stats PlayerStats
	var lastGamble sql.NullTime

	err := r.db.QueryRow(queries.GetPlayerStats, playerID).Scan(
		&stats.PlayerID,
		&stats.TotalGambles,
		&stats.TotalWagered,
		&stats.TotalWon,
		&stats.TotalLost,
		&stats.Wins,
		&stats.Losses,
		&lastGamble,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if lastGamble.Valid {
		stats.LastGamble = &lastGamble.Time
	}

	return &stats, nil
}

func (r *playerStatsRepository) RecordWin(playerID int, wager int, won int) error {
	_, err := r.db.Exec(
		queries.UpdateStatsWin,
		wager,
		won,
		playerID,
	)
	return err
}

func (r *playerStatsRepository) RecordLoss(playerID int, wager int, lost int) error {
	_, err := r.db.Exec(
		queries.UpdateStatsLoss,
		wager,
		lost,
		playerID,
	)
	return err
}

func (r *playerStatsRepository) Reset(playerID int) error {
	_, err := r.db.Exec(queries.ResetPlayerStats, playerID)
	return err
}

func (r *walletStatsRepository) Create(playerID int) error {
	_, err := r.db.Exec(queries.CreateWalletStats, playerID)
	return err
}

func (r *walletStatsRepository) Get(playerID int) (*WalletStats, error) {
	var stats WalletStats

	err := r.db.QueryRow(queries.GetWalletStats, playerID).Scan(
		&stats.PlayerID,
		&stats.Balance,
		&stats.TotalPaid,
		&stats.TotalReceived,
		&stats.DepositCount,
		&stats.WithdrawCount,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &stats, nil
}

func (r *walletStatsRepository) Deposit(playerID int, amount int) error {
	_, err := r.db.Exec(queries.WalletDeposit, amount, playerID)
	return err
}

func (r *walletStatsRepository) Withdraw(playerID int, amount int) error {
	_, err := r.db.Exec(queries.WalletWithdraw, amount, playerID)
	return err
}

func (r *walletStatsRepository) Pay(playerID int, amount int) error {
	_, err := r.db.Exec(queries.WalletPay, amount, playerID)
	return err
}

func (r *walletStatsRepository) Receive(playerID int, amount int) error {
	_, err := r.db.Exec(queries.WalletReceive, amount, playerID)
	return err
}

func (r *walletStatsRepository) Reset(playerID int) error {
	_, err := r.db.Exec(queries.ResetWalletStats, playerID)
	return err
}

func (r *gamblingStatsRepository) Init() error {
	_, err := r.db.Exec(queries.InitGlobalStats)
	return err
}

func (r *gamblingStatsRepository) Get() (*GamblingStats, error) {
	var stats GamblingStats

	err := r.db.QueryRow(queries.GetGlobalStats).Scan(
		&stats.TotalGambles,
		&stats.TotalWagered,
		&stats.TotalPaid,
		&stats.LastUpdate,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &stats, nil
}

func (r *gamblingStatsRepository) UpdateAfterGamble(wager int, paid int) error {
	_, err := r.db.Exec(
		queries.UpdateGlobalAfterGamble,
		wager,
		paid,
	)
	return err
}

func (r *gamblingStatsRepository) Reset() error {
	_, err := r.db.Exec(queries.ResetGlobalStats)
	return err
}
