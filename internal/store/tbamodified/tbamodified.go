package tbamodified

// TBAModified holds information about when a TBA endpoint was last modified.
type TBAModified struct {
	Endpoint     string
	LastModified string
}

// Service provides an interface for interacting with a store for when tba requests were last modified.
type Service interface {
	EventsModified() (string, error)
	SetEventsModified(lastModified string) error
	SetMatchModified(eventKey string, lastModified string) error
	MatchModified(eventKey string) (string, error)
	Close() error
}
