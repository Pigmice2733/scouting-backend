CREATE TABLE IF NOT EXISTS reports (
	reporter TEXT NOT NULL,
	eventKey TEXT NOT NULL,
	matchKey TEXT NOT NULL,
	isBlue BOOLEAN NOT NULL,
	team TEXT NOT NULL,
	stats TEXT NOT NULL,
	UNIQUE(eventKey, matchKey, team),
  FOREIGN KEY(reporter) REFERENCES users(username),
	FOREIGN KEY(eventKey) REFERENCES events(key),
	FOREIGN KEY(matchKey) REFERENCES matches(key)
)