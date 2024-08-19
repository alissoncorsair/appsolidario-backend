package user

import (
	"database/sql"
	"fmt"

	"github.com/alissoncorsair/appsolidario-backend/service/profile_picture"
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
		INSERT INTO users (name, surname, email, password, status, description, postal_code, city, state, cpf, role_id, points, birth_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`
	var id int
	err := s.db.QueryRow(query, user.Name, user.Surname, user.Email, user.Password, defaultStatus, user.Description, user.PostalCode, user.City, user.State, user.CPF, user.RoleID, defaultPoints, user.BirthDate).Scan(&id)

	if err != nil {
		return nil, err
	}

	user.ID = id
	return user, nil
}

func (s *Store) GetUserByEmail(email string) (*types.User, error) {
	var u *types.User
	query := `
	SELECT id, name, surname, email, password, status, description, postal_code, city, state, cpf, role_id, points, birth_date, created_at, updated_at
	FROM users
	WHERE email = $1
	`
	row := s.db.QueryRow(query, email)
	u, err := ScanRowIntoUser(row)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) GetUserByCPF(cpf string) (*types.User, error) {
	var u *types.User
	query := `
	SELECT id, name, surname, email, password, status, description, postal_code, city, state, cpf, role_id, points, birth_date, created_at, updated_at
	FROM users
	WHERE cpf = $1
	`
	row := s.db.QueryRow(query, cpf)
	u, err := ScanRowIntoUser(row)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) GetUserByID(id int) (*types.User, error) {
	var u *types.User
	query := `
	SELECT id, name, surname, email, password, status, description, postal_code, city, state, cpf, role_id, points, birth_date, created_at, updated_at
	FROM users
	WHERE id = $1
	`
	row := s.db.QueryRow(query, id)
	u, err := ScanRowIntoUser(row)

	if err != nil {
		return nil, err
	}

	return u, nil
}

func (s *Store) GetUserProfilePicture(userID int) (*types.ProfilePicture, error) {
	var pp *types.ProfilePicture
	query := `
	SELECT id, user_id, path, created_at, updated_at
	FROM profile_pictures
	WHERE user_id = $1
	`
	row := s.db.QueryRow(query, userID)

	pp, err := profile_picture.ScanRowIntoProfilePicture(row)

	if err != nil {
		return nil, err
	}

	return pp, nil
}

func (s *Store) UpdateUserStatus(userID int, status types.UserStatus) error {
	query := `
	UPDATE users
	SET status = $1, updated_at = NOW()
	WHERE id = $2
	`
	result, err := s.db.Exec(query, status, userID)
	if err != nil {
		return fmt.Errorf("error updating user status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func ScanRowIntoUser(row *sql.Row) (*types.User, error) {
	var u types.User
	err := row.Scan(&u.ID, &u.Name, &u.Surname, &u.Email, &u.Password, &u.Status, &u.Description, &u.PostalCode, &u.City, &u.State, &u.CPF, &u.RoleID, &u.Points, &u.BirthDate, &u.CreatedAt, &u.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("error scanning user: %w", err)
	}

	return &u, nil
}
