package alliance

// Alliance holds all the team numbers in a given alliance.
type Alliance []string

// Service is a store for alliances.
type Service interface {
	Get(matchKey string, isBlue bool) (Alliance, error)
	Upsert(matchKey string, isBlue bool, alliance Alliance) error
}
