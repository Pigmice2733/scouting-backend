CREATE TABLE IF NOT EXISTS matches (
	key TEXT PRIMARY KEY,
	eventKey TEXT NOT NULL,
	predictedTime TIMESTAMPTZ,
	actualTime TIMESTAMPTZ,
	blueWon BOOLEAN,
	redScore INTEGER,
	blueScore INTEGER,
	FOREIGN KEY(eventKey) REFERENCES events(key)
)