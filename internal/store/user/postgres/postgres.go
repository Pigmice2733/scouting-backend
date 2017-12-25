package postgres

import "database/sql"
import "github.com/Pigmice2733/scouting-backend/internal/store/user"
import "github.com/Pigmice2733/scouting-backend/internal/store"

// Service is used for getting information about a user from a postgres database.
type Service struct {
	db *sql.DB
}

// New creates a new user service.
func New(db *sql.DB) user.Service {
	return &Service{db: db}
}

// Get gets a user with a given username from the postgresql database.
func (s *Service) Get(username string) (u user.User, err error) {
	err = s.db.QueryRow("SELECT username, hashedPassword FROM users WHERE username = $1", username).Scan(&u.Username, &u.HashedPassword)
	if err == sql.ErrNoRows {
		err = store.ErrNoResults
	}
	return
}

// Create creates a new user in the postgresql database.
func (s *Service) Create(u user.User) error {
	_, err := s.db.Exec("INSERT INTO users VALUES ($1, $2)", u.Username, u.HashedPassword)
	return err
}

// Delete removes an existing user with a given username from the postgresql database.
func (s *Service) Delete(username string) error {
	_, err := s.db.Exec("DELETE FROM users WHERE username = $1", username)
	return err
}
