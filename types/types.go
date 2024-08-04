package types

import "time"

// Config holds the configuration details for the database connection.
type Config struct {
	Host                          string
	Port                          int
	PostgresPassword              string
	PostgresUser                  string
	PostgresDB                    string
	SSLMode                       string
	JWTExpirationInSeconds        int64
	JWTRefreshExpirationInSeconds int64
	JWTSecret                     string
}

// UserRole defines the role of a user.
type UserRole int

const (
	RolePayee UserRole = 1
	RolePayer UserRole = 2
)

// UserStatus defines the status of a user.
type UserStatus int

const (
	StatusInactive UserStatus = 0
	StatusActive   UserStatus = 1
)

// Role represents the role entity.
type Role struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// User represents the user entity.
type UserWithoutPassword struct {
	ID         int    `json:"id"`
	Name       string `json:"name" validate:"required,min=3,max=100"`
	Surname    string `json:"surname" validate:"required,min=3,max=100"`
	Email      string `json:"email" validate:"required,email"`
	PostalCode string `json:"postal_code" validate:"required,len=8"`
	City       string `json:"city" validate:"required,max=100"`
	// Street           string     `json:"street" validate:"required,max=255"`
	State       string     `json:"state" validate:"required,max=100"`
	Status      UserStatus `json:"status" validate:"required,oneof=0 1"` // 0 for inactive, 1 for active
	Description *string    `json:"description,omitempty" validate:"omitempty,max=1000"`
	CPF         string     `json:"cpf" validate:"required,len=11"`
	RoleID      UserRole   `json:"role_id" validate:"required"`
	Points      int        `json:"points" validate:"gte=0"`
	BirthDate   time.Time  `json:"birth_date" validate:"required"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type User struct {
	UserWithoutPassword
	Password string `json:"password"`
}

type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	GetUserByID(id int) (*User, error)
	CreateUser(user *User) (*User, error)
}

// ProfilePicture represents the profile picture entity.
type ProfilePicture struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RegisterUserRequest struct {
	Name        string `json:"name" validate:"required,min=3,max=100"`
	Surname     string `json:"surname" validate:"required,min=3,max=100"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required,min=6"`
	Description string `json:"description" validate:"omitempty,max=1000"`
	PostalCode  string `json:"postal_code" validate:"required,len=8"`
	City        string `json:"city" validate:"required,max=100"`
	State       string `json:"state" validate:"required,max=100"`
	CPF         string `json:"cpf" validate:"required,min=11,max=14"`
	RoleID      int    `json:"role_id" validate:"required,oneof=1 2"`
	BirthDate   string `json:"birth_date" validate:"required"`
}

type LoginUserRequest struct {
	CPF      string `json:"cpf" validate:"required,min=11,max=14"`
	Password string `json:"password" validate:"required,min=6"`
}
