package token

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

func (s *Store) CreateToken(token *types.Token) (*types.Token, error) {
	query := `
		INSERT INTO tokens (user_id, token, token_type)
		VALUES ($1, $2, $3)
		RETURNING id
	`
	var id int
	err := s.db.QueryRow(query, token.UserID, token.Token, token.Type).Scan(&id)

	if err != nil {
		return nil, err
	}

	token.ID = id
	return token, nil
}

func (s *Store) GetToken(token string) (*types.Token, error) {
	query := `
		SELECT * FROM tokens WHERE token = $1
	`

	row := s.db.QueryRow(query, token)
	return ScanRowIntoToken(row)
}

func ScanRowIntoToken(row *sql.Row) (*types.Token, error) {
	var t types.Token
	err := row.Scan(&t.ID, &t.UserID, &t.Token, &t.Type, &t.CreatedAt, &t.ExpiresAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &t, nil
}
