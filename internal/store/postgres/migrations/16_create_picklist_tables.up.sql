CREATE TABLE IF NOT EXISTS picklists (
    id SERIAL PRIMARY KEY,
    eventKey TEXT NOT NULL,
    name TEXT,
    owner TEXT,
    FOREIGN KEY(eventKey) REFERENCES events(key),
    FOREIGN KEY(owner) REFERENCES users(username)
);

CREATE TABLE IF NOT EXISTS picks (
    picklistId integer NOT NULL,
    team TEXT NOT NULL,
    FOREIGN KEY(picklistId) REFERENCES picklists(id)
);