package bank

import "fmt"

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Init(initialBalance int) error {
	balance, err := s.repo.Balance()
	if err != nil {
		return err
	}

	if balance != 0 {
		return nil
	}

	return s.repo.Create(initialBalance)
}

func (s *Service) Balance() (int, error) {
	return s.repo.Balance()
}

func (s *Service) SetBalance(amount int) error {
	return s.repo.SetBalance(amount)
}

func (s *Service) Deposit(amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.repo.Deposit(amount)
}

func (s *Service) Withdraw(amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}
	return s.repo.Withdraw(amount)
}
