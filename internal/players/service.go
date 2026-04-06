package players

import (
	"fmt"
	"plugin/internal/config"
	"plugin/internal/iw4m"
)

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(name, xuid, guid string, level int, cfg *config.Config, iw4m *iw4m.IW4MWrapper) (int, error) {
	if exists, err := s.repo.ExistsByXUID(xuid); err != nil {
		return 0, fmt.Errorf("check xuid exists: %w", err)
	} else if exists {
		return 0, fmt.Errorf("player %s (%s) already exists", name, xuid)
	}

	if exists, err := s.repo.ExistsByGUID(guid); err != nil {
		return 0, fmt.Errorf("check guid exists: %w", err)
	} else if exists {
		return 0, fmt.Errorf("player %s (%s) already exists", name, guid)
	}

	var iw4mID *int
	if cfg.IW4MAdmin.Enabled && iw4m != nil {
		iw4mID = iw4m.ClientIDFromGUID(guid)
	}

	id, err := s.repo.Create(name, xuid, guid, level, iw4mID)
	if err != nil {
		return 0, fmt.Errorf("create player: %w", err)
	}

	return id, nil
}

func (s *Service) GetByID(id int) (*Player, error) {
	return s.repo.GetByID(id)
}

func (s *Service) GetByXUID(xuid string) (*Player, error) {
	return s.repo.GetByXUID(xuid)
}

func (s *Service) GetByGUID(guid string) (*Player, error) {
	return s.repo.GetByGUID(guid)
}

func (s *Service) GetByLevel(level int) (*Player, error) {
	return s.repo.GetByLevel(level)
}

func (s *Service) ExistsByLevel(level int) (bool, error) {
	return s.repo.ExistsByLevel(level)
}

func (s *Service) GetByDiscordID(id string) (*Player, error) {
	return s.repo.GetByDiscordID(id)
}

func (s *Service) GetDiscordIDByID(id int) (string, error) {
	d, err := s.repo.GetDiscordIDByID(id)
	if err != nil {
		return "", fmt.Errorf("get discord id: %w", err)
	}
	if d == "" {
		return "", fmt.Errorf("player has no discord id")
	}

	return d, nil
}

func (s *Service) GetByPartial(partial string) (*Player, error) {
	return s.repo.GetByPartial(partial)
}

func (s *Service) UpdateDiscordID(id int, discordID string) error {
	if discordID == "" {
		return fmt.Errorf("discordID cannot be empty")
	}
	return s.repo.UpdateDiscordID(id, discordID)
}

func (s *Service) UpdateName(id int, name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}
	return s.repo.UpdateName(id, name)
}

func (s *Service) UpdateLevel(id int, level int) error {
	if level < 0 {
		return fmt.Errorf("level cannot be negative")
	}
	return s.repo.UpdateLevel(id, level)
}

func (s *Service) Delete(id int) error {
	return s.repo.Delete(id)
}

func (s *Service) ExistsByID(id int) (bool, error) {
	return s.repo.ExistsByID(id)
}

func (s *Service) ExistsByXUID(xuid string) (bool, error) {
	return s.repo.ExistsByXUID(xuid)
}

func (s *Service) ExistsByGUID(guid string) (bool, error) {
	return s.repo.ExistsByGUID(guid)
}

func (s *Service) GetAllPlayers() ([]*Player, error) {
	return s.repo.GetAll()
}
