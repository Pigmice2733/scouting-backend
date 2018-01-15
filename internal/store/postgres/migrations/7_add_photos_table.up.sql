CREATE TABLE IF NOT EXISTS photos (
	team TEXT NOT NULL,
	url TEXT NOT NULL,
    UNIQUE(team, url)
)