package stats

import "fmt"

type PlayerStatsService struct {
	repo PlayerStatsRepository
}

func NewPlayerStats(repo PlayerStatsRepository) *PlayerStatsService {
	return &PlayerStatsService{repo: repo}
}

type WalletStatsService struct {
	repo WalletStatsRepository
}

func NewWalletStats(repo WalletStatsRepository) *WalletStatsService {
	return &WalletStatsService{repo: repo}
}

type GamblingStatsService struct {
	repo GamblingStatsRepository
}

func NewGamblingStats(repo GamblingStatsRepository) *GamblingStatsService {
	return &GamblingStatsService{repo: repo}
}

func (s *PlayerStatsService) Init(playerID int) error {
	stats, err := s.repo.Get(playerID)
	if err != nil {
		return err
	}
	if stats != nil {
		return nil
	}
	return s.repo.Create(playerID)
}

func (s *PlayerStatsService) GetStats(playerID int) (*PlayerStats, error) {
	return s.repo.Get(playerID)
}

func (s *PlayerStatsService) Win(playerID int, wager int, payout int) error {
	if wager <= 0 || payout <= 0 {
		return fmt.Errorf("invalid wager or payout")
	}
	return s.repo.RecordWin(playerID, wager, payout)
}

func (s *PlayerStatsService) Loss(playerID int, wager int) error {
	if wager <= 0 {
		return fmt.Errorf("invalid wager")
	}
	return s.repo.RecordLoss(playerID, wager, wager)
}

func (s *PlayerStatsService) Reset(playerID int) error {
	return s.repo.Reset(playerID)
}

func (s *WalletStatsService) Init(playerID int) error {
	stats, err := s.repo.Get(playerID)
	if err != nil {
		return err
	}
	if stats != nil {
		return nil
	}
	return s.repo.Create(playerID)
}

func (s *WalletStatsService) GetStats(playerID int) (*WalletStats, error) {
	return s.repo.Get(playerID)
}

func (s *WalletStatsService) Deposit(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("invalid deposit amount")
	}
	return s.repo.Deposit(playerID, amount)
}

func (s *WalletStatsService) Withdraw(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("invalid withdraw amount")
	}
	return s.repo.Withdraw(playerID, amount)
}

func (s *WalletStatsService) Pay(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("invalid payout amount")
	}
	return s.repo.Pay(playerID, amount)
}

func (s *WalletStatsService) Receive(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("invalid received amount")
	}
	return s.repo.Receive(playerID, amount)
}

func (s *WalletStatsService) Reset(playerID int) error {
	return s.repo.Reset(playerID)
}

func (s *GamblingStatsService) Init() error {
	stats, err := s.repo.Get()
	if err != nil {
		return err
	}

	if stats != nil {
		return nil
	}

	return s.repo.Init()
}

func (s *GamblingStatsService) GetStats() (*GamblingStats, error) {
	return s.repo.Get()
}

func (s *GamblingStatsService) RecordGamble(wager int, paid int) error {
	if wager <= 0 {
		return fmt.Errorf("wager must be positive")
	}
	if paid < 0 {
		return fmt.Errorf("paid cannot be negative")
	}
	return s.repo.UpdateAfterGamble(wager, paid)
}

func (s *GamblingStatsService) Reset() error {
	return s.repo.Reset()
}
