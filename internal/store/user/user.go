package user

// User holds information about a user.
type User struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashedPassword"`
	IsAdmin        bool   `json:"isAdmin"`
}

// Service is a store for users.
type Service interface {
	Create(User) error
	Get(username string) (User, error)
	GetUsers() ([]User, error)
	Update(username string, u User) error
	Delete(username string) error
}
