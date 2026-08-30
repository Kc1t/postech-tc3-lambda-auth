package customer

import (
	"context"
	"database/sql"
	"errors"
)

var (
	ErrNotFound = errors.New("cliente nao encontrado")
	ErrInactive = errors.New("cliente inativo")
)

type Customer struct {
	ID     string
	Name   string
	Active bool
	Role   string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByDocument(ctx context.Context, document string) (Customer, error) {
	const query = `SELECT id, name, active, role FROM customers WHERE document = $1`

	var c Customer

	err := r.db.QueryRowContext(ctx, query, document).Scan(&c.ID, &c.Name, &c.Active, &c.Role)
	if errors.Is(err, sql.ErrNoRows) {
		return Customer{}, ErrNotFound
	}

	if err != nil {
		return Customer{}, err
	}

	if !c.Active {
		return Customer{}, ErrInactive
	}

	return c, nil
}
