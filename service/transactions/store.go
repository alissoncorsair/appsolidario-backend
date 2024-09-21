package transactions

import (
	"database/sql"

	"github.com/alissoncorsair/appsolidario-backend/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func ScanRowIntoTransaction(row *sql.Row) (*types.Transaction, error) {
	var t types.Transaction
	err := row.Scan(&t.ID, &t.ExternalID, &t.PayerID, &t.PayeeID, &t.Amount, &t.Status, &t.Description, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *Store) CreateTransaction(externalId string, payerID, payeeID int, amount float64, description string) (*types.Transaction, error) {
	transaction, err := s.GetTransactionByExternalID(externalId)

	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if transaction != nil {
		return nil, nil
	}

	query := `INSERT INTO transactions (external_id, payer_id, payee_id, amount, status, description) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, external_id, payer_id, payee_id, amount, status, description, created_at, updated_at`
	row := s.db.QueryRow(query, externalId, payerID, payeeID, amount, types.StatusPending, description)

	return ScanRowIntoTransaction(row)
}

func (s *Store) UpdateTransactionStatus(externalId string, status types.TransactionStatus) (*types.Transaction, error) {
	query := `UPDATE transactions SET status = $1 WHERE external_id = $2 RETURNING id, external_id, payer_id, payee_id, amount, status, description, created_at, updated_at`
	row := s.db.QueryRow(query, status, externalId)

	return ScanRowIntoTransaction(row)
}

func (s *Store) GetTransactionByExternalID(externalID string) (*types.Transaction, error) {
	query := `SELECT id, external_id, payer_id, payee_id, amount, status, description, created_at, updated_at FROM transactions WHERE external_id = $1`
	row := s.db.QueryRow(query, externalID)

	return ScanRowIntoTransaction(row)
}
