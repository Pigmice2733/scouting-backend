package user

// User holds the credentials for a  user
type User struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashedPassword"`
}

// Service provides an interface for interacting with a store for users
type Service interface {
	Create(User) error
	Get(username string) (User, error)
	GetUsers() ([]User, error)
	Delete(username string) error
	Close() error
}
