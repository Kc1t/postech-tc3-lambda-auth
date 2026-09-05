package requester

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound = errors.New("cliente nao encontrado")
	ErrInactive = errors.New("cliente inativo")
)

const statusActive = "active"

type Requester struct {
	ID     string
	Name   string
	Email  string
	Status string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByDocument(ctx context.Context, document string) (Requester, error) {
	const query = `SELECT id, name, email, status FROM requesters WHERE document = $1`

	var found Requester

	err := r.db.QueryRowContext(ctx, query, document).Scan(&found.ID, &found.Name, &found.Email, &found.Status)
	if errors.Is(err, sql.ErrNoRows) {
		return Requester{}, ErrNotFound
	}

	if err != nil {
		return Requester{}, err
	}

	if found.Status != statusActive {
		return found, ErrInactive
	}

	return found, nil
}
