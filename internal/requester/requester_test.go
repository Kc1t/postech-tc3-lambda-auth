package requester_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/Kc1t/postech-tc3-lambda-auth/internal/requester"
)

const query = `SELECT id, name, email, status FROM requesters WHERE document = \$1`

func TestFindByDocument_Active(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("falha ao criar o mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "status"}).
		AddRow("req-1", "João Silva", "joao@email.com", "active")
	mock.ExpectQuery(query).WithArgs("52998224725").WillReturnRows(rows)

	found, err := requester.NewRepository(db).FindByDocument(context.Background(), "52998224725")
	if err != nil {
		t.Fatalf("esperava sucesso, veio %v", err)
	}

	if found.ID != "req-1" || found.Name != "João Silva" || found.Email != "joao@email.com" {
		t.Errorf("dados inesperados: %+v", found)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("expectativas nao atendidas: %v", err)
	}
}

func TestFindByDocument_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("falha ao criar o mock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(query).WithArgs("52998224725").WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "status"}))

	if _, err := requester.NewRepository(db).FindByDocument(context.Background(), "52998224725"); !errors.Is(err, requester.ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, veio %v", err)
	}
}

func TestFindByDocument_Inactive(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("falha ao criar o mock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "name", "email", "status"}).
		AddRow("req-1", "João Silva", "joao@email.com", "inactive")
	mock.ExpectQuery(query).WithArgs("52998224725").WillReturnRows(rows)

	found, err := requester.NewRepository(db).FindByDocument(context.Background(), "52998224725")
	if !errors.Is(err, requester.ErrInactive) {
		t.Fatalf("esperava ErrInactive, veio %v", err)
	}

	if found.ID != "req-1" {
		t.Errorf("esperava o solicitante devolvido junto do erro, veio %+v", found)
	}
}

func TestFindByDocument_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("falha ao criar o mock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(query).WithArgs("52998224725").WillReturnError(errors.New("conexao caiu"))

	_, err = requester.NewRepository(db).FindByDocument(context.Background(), "52998224725")
	if err == nil || errors.Is(err, requester.ErrNotFound) || errors.Is(err, requester.ErrInactive) {
		t.Fatalf("esperava o erro cru do banco, veio %v", err)
	}
}
