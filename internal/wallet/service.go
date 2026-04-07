package wallet

import "fmt"

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(id, startingBalance int) error {
	exists, err := s.repo.Exists(id)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("wallet already exists (%d)", id)
	}

	return s.repo.Create(id, startingBalance)
}

func (s *Service) Balance(playerID int) (int, error) {
	return s.repo.Balance(playerID)
}

func (s *Service) SetBalance(playerID int, amount int) error {
	return s.repo.SetBalance(playerID, amount)
}

func (s *Service) Deposit(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("deposit amount must be positive")
	}
	return s.repo.Deposit(playerID, amount)
}

func (s *Service) Withdraw(playerID int, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("withdraw amount must be positive")
	}

	return s.repo.Withdraw(playerID, amount)
}

func (s *Service) DeleteWallet(playerID int) error {
	return s.repo.Delete(playerID)
}

func (s *Service) Exists(playerID int) (bool, error) {
	return s.repo.Exists(playerID)
}

func (s *Service) GetTop5RichestWallets() ([]wallet, error) {
	return s.repo.GetTopWallets(5)
}

func (s *Service) GetTop5PoorestWallets() ([]wallet, error) {
	return s.repo.GetBottomWallets(5)
}

func (s *Service) GetTop10RichestWallets() ([]wallet, error) {
	return s.repo.GetTopWallets(10)
}

func (s *Service) GetTop10PoorestWallets() ([]wallet, error) {
	return s.repo.GetBottomWallets(10)
}

func (s *Service) WalletToWallet(fromID, toID, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	if fromID == toID {
		return fmt.Errorf("cannot transfer to the same wallet")
	}

	bal, err := s.repo.Balance(fromID)
	if err != nil {
		return err
	}

	if bal < amount {
		return fmt.Errorf("insufficient balance")
	}

	if err := s.repo.Withdraw(fromID, amount); err != nil {
		return err
	}

	// Deposit to receiver
	if err := s.repo.Deposit(toID, amount); err != nil {
		// rollback attempt (very important)
		_ = s.repo.Deposit(fromID, amount)
		return fmt.Errorf("transfer failed (rolled back): %w", err)
	}

	return nil
}
