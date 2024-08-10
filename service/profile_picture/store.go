package profile_picture

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

func ScanRowIntoProfilePicture(row *sql.Row) (*types.ProfilePicture, error) {
	var pp types.ProfilePicture
	err := row.Scan(&pp.ID, &pp.UserID, &pp.Path, &pp.CreatedAt, &pp.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &pp, nil
}
