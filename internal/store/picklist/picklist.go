package picklist

// BasicPicklist defines a picklist with only it's ID, eventKey, and name.
type BasicPicklist struct {
	ID       int    `json:"id"`
	EventKey string `json:"eventKey"`
	Name     string `json:"name"`
}

// Picklist holds all information about a picklist.
type Picklist struct {
	BasicPicklist
	List  []string `json:"list"`
	Owner string   `json:"owner"`
}

// Service is a store for picklists.
type Service interface {
	GetBasicPicklists(username string) (bPicklists []BasicPicklist, err error)
	Get(id int) (p Picklist, err error)
	Insert(p Picklist) (id int, err error)
	Update(p Picklist) error
	GetOwner(id int) (username string, err error)
	GetByEvent(username, eventKey string) (bPicklists []BasicPicklist, err error)
}
