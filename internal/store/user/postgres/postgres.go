package postgres

import (
	"database/sql"

	"github.com/Pigmice2733/scouting-backend/internal/store"
	"github.com/Pigmice2733/scouting-backend/internal/store/user"
)

// Service holds a db for users.
type Service struct {
	db *sql.DB
}

// New returns a new Service with the given db.
func New(db *sql.DB) user.Service {
	return &Service{db: db}
}

// Close closes the postgresql db connection.
func (s *Service) Close() error {
	return s.db.Close()
}

// Get retrieves a user from the postgresql db.
func (s *Service) Get(username string) (user user.User, err error) {
	err = s.db.QueryRow("SELECT username, hashedPassword FROM users WHERE username = $1", username).Scan(&user.Username, &user.HashedPassword)
	if err == sql.ErrNoRows {
		err = store.ErrNoResults
	}
	return
}

// GetUsers retrieves all users in the postgresql db.
func (s *Service) GetUsers() ([]user.User, error) {
	var users []user.User

	rows, err := s.db.Query("SELECT username, hashedPassword FROM users")
	if err != nil {
		return users, err
	}
	defer rows.Close()

	for rows.Next() {
		var user user.User
		if err := rows.Scan(&user.Username, &user.HashedPassword); err != nil {
			return users, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// Create creates a user in the postgresql db.
func (s *Service) Create(user user.User) error {
	_, err := s.db.Exec("INSERT INTO users VALUES ($1, $2)", user.Username, user.HashedPassword)
	return err
}

// Delete deletes a user from the postgresql db.
func (s *Service) Delete(username string) error {
	_, err := s.db.Exec("DELETE FROM users WHERE username = $1", username)
	return err
}
