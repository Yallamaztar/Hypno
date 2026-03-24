package links

import (
	"database/sql"
	"errors"
	"plugin/internal/database/queries"
	"time"
)

type Repository interface {
	Create(id int, code string, expiresAt time.Time) error
	GetIDByCode(code string) (int, error)
	GetCodeByID(id int) (string, error)
	DeleteByID(id int) error
	DeleteByCode(code string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(id int, code string, expiresAt time.Time) error {
	if err := r.DeleteByID(id); err != nil {
		return err
	}

	_, err := r.db.Exec(queries.CreateLink, id, code, expiresAt)
	return err
}

func (r *repository) GetIDByCode(code string) (int, error) {
	var id int
	err := r.db.QueryRow(
		queries.GetIDByCode,
		code,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, err
	}

	return id, nil
}

func (r *repository) GetCodeByID(id int) (string, error) {
	var code string
	err := r.db.QueryRow(
		queries.GetCodeByID,
		id,
	).Scan(&code)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", err
	}

	return code, nil
}

func (r *repository) DeleteByID(id int) error {
	_, err := r.db.Exec(
		queries.DeleteLinkByPlayer,
		id,
	)
	return err
}

func (r *repository) DeleteByCode(code string) error {
	_, err := r.db.Exec(
		queries.DeleteLinkByCode,
		code,
	)
	return err
}
