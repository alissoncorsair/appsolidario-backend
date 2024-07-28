package types

type Config struct {
	Host             string
	Port             int
	PostgresPassword string
	PostgresUser     string
	PostgresDB       string
	SSLMode          string
}
