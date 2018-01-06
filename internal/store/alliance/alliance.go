package alliance

// Alliance holds all the team numbers in a given alliance.
type Alliance []string

// Service is a store for alliances.
type Service interface {
	GetColor(matchKey string, number string) (bool, error)
	Get(matchKey string, isBlue bool) (Alliance, error)
	Upsert(matchKey string, isBlue bool, alliance Alliance) error
}
