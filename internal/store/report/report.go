package report

// Auto holds information about the autonomous performance in a match
type Auto struct {
	CrossedLine   bool `json:"crossedLine"`
	DeliveredGear bool `json:"deliveredGear"`
	Fuel          int  `json:"fuel"`
}

// Teleop holds data about how a team performed in the teleop section of a match
type Teleop struct {
	Climbed bool `json:"climbed"`
	Gears   int  `json:"gears"`
	Fuel    int  `json:"fuel"`
}

// Report holds information about a team and their the performance in a specific match
type Report struct {
	Reporter string `json:"reporter"`
	Alliance string `json:"alliance"`
	Team     string `json:"team"`
	Score    int    `json:"score"`
	Auto     Auto   `json:"auto"`
	Teleop   Teleop `json:"teleop"`
}

// Service provides an interface for interacting with a store for reports
type Service interface {
	Create(rd Report, allianceID int) error
	Update(rd Report, allianceID int) error
	Close() error
}
