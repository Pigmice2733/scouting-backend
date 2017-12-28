CREATE TABLE IF NOT EXISTS alliances (
	matchKey TEXT NOT NULL,
	isBlue BOOLEAN NOT NULL,
	number TEXT NOT NULL,
	FOREIGN KEY(matchKey) REFERENCES matches(key),
	UNIQUE(matchKey, number)
)