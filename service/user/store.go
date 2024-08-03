package user

import (
	"database/sql"

	"github.com/alissoncorsair/appsolidario-backend/types"
)

// Store represents the store for user.
type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		db: db,
	}
}

func (s *Store) CreateUser(user *types.User) (*types.User, error) {
	defaultStatus := types.StatusInactive
	defaultPoints := 0
	query := `
		INSERT INTO users (name, email, password, status, description, postal_code, city, state, cpf, role_id, points, registration_date, birth_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id
	`
	var id int
	err := s.db.QueryRow(query, user.Name, user.Email, user.Password, defaultStatus, user.Description, user.PostalCode, user.City, user.State, user.CPF, user.RoleID, defaultPoints, user.RegistrationDate, user.BirthDate).Scan(&id)

	if err != nil {
		return nil, err
	}

	user.ID = id
	return user, nil
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	var u types.User
	query := `
	SELECT id, name, email, password, status, cpf, role_id, points
	FROM users
	WHERE email = $1
`
	err := s.db.QueryRow(query, email).Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Status, &u.CPF, &u.RoleID, &u.Points)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
