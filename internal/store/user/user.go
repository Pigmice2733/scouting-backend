package user

// User holds information about a user.
type User struct {
	Username       string `json:"username"`
	HashedPassword string `json:"hashedPassowrd"`
	IsAdmin        bool   `json:"isAdmin"`
}

// Service is a store for users.
type Service interface {
	Get(username string) (User, error)
	Create(User) error
	Delete(username string) error
}
