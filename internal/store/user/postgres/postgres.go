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

// Create creates a new user in the postgresql database.
func (s *Service) Create(u user.User) error {
	_, err := s.db.Exec("INSERT INTO users VALUES ($1, $2, $3, $4)", u.Username, u.HashedPassword, u.IsAdmin, u.IsVerified)
	return err
}

// Get gets a user with a given username from the postgresql database.
func (s *Service) Get(username string) (u user.User, err error) {
	err = s.db.QueryRow("SELECT username, hashedPassword, isAdmin, isVerified FROM users WHERE username = $1", username).Scan(&u.Username, &u.HashedPassword, &u.IsAdmin, &u.IsVerified)
	if err == sql.ErrNoRows {
		err = store.ErrNoResults
	}
	return
}

// GetUsers gets all users in the postgresql database.
func (s *Service) GetUsers() ([]user.User, error) {
	rows, err := s.db.Query("SELECT username, hashedPassword, isAdmin, isVerified from users")
	if err != nil {
		return []user.User{}, err
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		var u user.User
		if err := rows.Scan(&u.Username, &u.HashedPassword, &u.IsAdmin, &u.IsVerified); err != nil {
			return users, err
		}
		users = append(users, u)
	}

	return users, rows.Err()
}

// Update updates a given user in the postgresql database.
func (s *Service) Update(username string, nu user.NullableUser) error {
	_, err := s.db.Exec(`
		UPDATE users
			SET
				username = COALESCE($1, username),
				hashedPassword = COALESCE($2, hashedPassword),
				isAdmin = COALESCE($3, isAdmin),
				isVerified = COALESCE($4, isVerified)
			WHERE
				username = $5
		`, nu.Username, nu.HashedPassword, nu.IsAdmin, nu.IsVerified, username)
	return err
}

// Delete removes an existing user with a given username from the postgresql database.
func (s *Service) Delete(username string) error {
	_, err := s.db.Exec("DELETE FROM users WHERE username = $1", username)
	return err
}
