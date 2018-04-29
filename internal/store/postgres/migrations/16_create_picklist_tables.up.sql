CREATE TABLE IF NOT EXISTS picklists (
    id uuid UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    eventKey TEXT NOT NULL,
    name TEXT,
    owner TEXT,
    FOREIGN KEY(eventKey) REFERENCES events(key),
    FOREIGN KEY(owner) REFERENCES users(username)
);

CREATE TABLE IF NOT EXISTS picks (
    picklistId uuid NOT NULL,
    team TEXT NOT NULL,
    FOREIGN KEY(picklistId) REFERENCES picklists(id)
);