package photo

// Service is a store for photos.
type Service interface {
	Get(team string) (string, error)
	Create(team, url string) error
}
