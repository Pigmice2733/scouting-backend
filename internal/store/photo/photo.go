package photo

// Service is a store for photos.
type Service interface {
	Exists(team string) (bool, error)
	Get(team string) (string, error)
	Create(team, url string) error
}
