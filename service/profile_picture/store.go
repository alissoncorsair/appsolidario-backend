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

func (s *Store) UpdateProfilePicture(pp *types.ProfilePicture) (*types.ProfilePicture, error) {
	query := `
		UPDATE profile_pictures
		SET path = $1
		WHERE id = $2
		RETURNING id, user_id, path, created_at, updated_at
	`
	pp, err := ScanRowIntoProfilePicture(s.db.QueryRow(query, pp.Path, pp.ID))

	if err != nil {
		return nil, err
	}

	return pp, nil
}

func (s *Store) CreateProfilePicture(pp *types.ProfilePicture) (*types.ProfilePicture, error) {
	query := `
		INSERT INTO profile_pictures (user_id, path)
		VALUES ($1, $2)
		RETURNING id, user_id, path, created_at, updated_at
	`
	pp, err := ScanRowIntoProfilePicture(s.db.QueryRow(query, pp.UserID, pp.Path))

	if err != nil {
		return nil, err
	}

	return pp, nil
}
