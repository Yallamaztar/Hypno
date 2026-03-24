package links

import "time"

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(id int, code string) error {
	expires := time.Now().Add(2 * time.Minute)
	return s.repo.Create(id, code, expires)
}

func (s *Service) GetIDByCode(code string) (int, error) {
	return s.repo.GetIDByCode(code)
}

func (s *Service) GetCodeByID(id int) (string, error) {
	return s.repo.GetCodeByID(id)
}

func (s *Service) DeleteByID(id int) error {
	return s.repo.DeleteByID(id)
}

func (s *Service) DeleteByCode(code string) error {
	return s.repo.DeleteByCode(code)
}
