# Data Types

## Events

### Go

| Name          | Type          | Tags | Example                         |
| ------------- | ------------- | ---- | ------------------------------- |
| Key           | string        |      | "2017cmptx"                     |
| Name          | string        |      | "Einstein Field (Houston)"      |
| Date          | time.Time     |      | 2017-04-21 17:00:00 -0700 -0700 |
| Matches       | []match.Match |      | See [Matches](#matches)         |

### PostgreSQL

| Name | Type             | Constraints | Example                         |
| ---- | ---------------- | ----------- | ------------------------------- |
| key  | TEXT PRIMARY KEY |             | "2017cmptx"                     |
| name | TEXT             | NOT NULL    | "Einstein Field (Houston)"      |
| date | TIMESTAMPTZ      | NOT NULL    | 2017-04-21 17:00:00 -0700 -0700 |

### JSON

| Name    | Type            | Comments | Example                    |
| ------- | ----------------| -------- | -------------------------- |
| key     | string          |          | "2017cmptx"                |
| name    | string          |          | "Einstein Field (Houston)" |
| date    | int (UNIX TIME) |          | 1512764281                 |
| matches | []match.Match   |          | See [Matches](#matches)    |


## Matches

### Go

| Name            | Type              | Tags      | Example                         |
| --------------- | ----------------- | --------- | ------------------------------- |
| Key             | string            |           | "2017cmptx_sf1m13"              |
| EventKey        | string            | -         | "2017cmptx"                     |
| PredictedTime   | time.Time         | omitempty | 2017-04-21 17:00:00 -0700 -0700 |
| ActualTime      | time.Time         | omitempty | 2017-04-21 17:00:00 -0700 -0700 |
| WinningAlliance | string            |           | "blue"                          |
| RedAlliance     | alliance.Alliance |           | See [Alliances](#alliances)     |
| BlueAlliance    | alliance.Alliance |           | See [Alliances](#alliances)     |

### PostgreSQL

| Name            | Type              | Constraints                                  | Example                         |
| --------------- | ----------------- | -------------------------------------------- | ------------------------------- |
| key             | TEXT PRIMARY KEY  |                                              | "2017cmptx_sf1m13"              |
| eventKey        | TEXT NOT NULL     | FOREIGN KEY(eventKey) REFERENCES events(key) | "2017cmptx"                     |
| predictedTime   | TIMESTAMPTZ       |                                              | 2017-04-21 17:00:00 -0700 -0700 |
| actualTime      | TIMESTAMPTZ       |                                              | 2017-04-21 17:00:00 -0700 -0700 |
| winningAlliance | TEXT              |                                              | "blue"                          |
| redAlliance     | alliance.Alliance |                                              | See [Alliances](#alliances)     |
| blueAlliance    | alliance.Alliance |                                              | See [Alliances](#alliances)     |

### JSON

| Name            | Type              | Comments      | Example                     |
| --------------- | ----------------- | ------------- | --------------------------- |
| key             | string            |               | "2017cmptx_sf1m13"          |
| predictedTime   | int (UNIX TIME)   | Omit if empty | 1512764281                  |
| actualTime      | int (UNIX TIME)   | Omit if empty | 1512764281                  |
| winningAlliance | string            |               | "blue"                        |
| redAlliance     | alliance.Alliance |               | See [Alliances](#alliances) |
| blueAlliance    | alliance.Alliance |               | See [Alliances](#alliances) |


## Alliances

### Go

| Name          | Type              | Tags      | Example                           |
| ------------- | ----------------- | --------- | --------------------------------- |
| MatchKey      | string            | -         | "2017cmptx_sf1m13"                |
| IsBlue        | bool              | -         | true                              |
| Score         | \*int             | omitempty | 96                                |
| Teams         | []string          |           | ["frc2471", "frc2733", "frc1418"] |

### PostgreSQL

| Name          | Type                       | Constraints                                   | Example                         |
| ------------- | -------------------------- | --------------------------------------------- | ------------------------------- |
| matchKey      | TEXT NOT NULL              | FOREIGN KEY(matchKey) REFERENCES matches(key) | "2017cmptx_sf1m13"              |
| isBlue        | BOOLEAN NOT NULL           |                                               | true                            |
| score         | INT NOT NULL               |                                               | 96                              |
|               |                            | UNIQUE (matchKey, isBlue)                     |                                 |

### JSON

| Name          | Type              | Comments      | Example                           |
| ------------- | ----------------- | ------------- | --------------------------------- |
| score         | int               | Omit if empty | 96                                |
| teams         | []string          |               | ["frc2471", "frc2733", "frc1418"] |


## Teams

### PostgreSQL

| Name          | Type                       | Constraints                                            | Example                         |
| ------------- | -------------------------- | ------------------------------------------------------ | ------------------------------- |
| matchKey      | TEXT                       | NOT NULL FOREIGN KEY(matchKey) REFERENCES matches(key) | "2017cmptx_sf1m13"              |
| isBlue        | BOOLEAN                    | NOT NULL                                               | true                            |
| number        | TEXT                       | NOT NULL                                               | "frc2733b"                      |


## TBA Modified

### Go

| Name         | Type   | Tags | Example                         |
| ------------ | ------ | ---- | ------------------------------- |
| Endpoint     | string |      | "/events/2017cc"                |
| LastModified | string |      | "Wed, 21 Oct 2015 07:28:00 GMT" |

### PostgreSQL

| Name                  | Type             | Constraints | Example                         |
| --------------------- | ---------------- | ----------- | ------------------------------- |
| endpoint              | TEXT PRIMARY KEY |             | "/events/2017cc"                |
| lastModified          | TEXT NOT NULL    |             | "Wed, 21 Oct 2015 07:28:00 GMT" |

### JSON

Not sent to front end. Used for time conditional requests to the TBA API. See [HTTP Last-Modified](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Last-Modified) and [HTTP If-Modified-Since](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/If-Modified-Since) for more information.


## Users

### Go

| Name           | Type   | Tags | Example        |
| -------------- | ------ | ---- | -------------- |
| Username       | string |      | "JohnSmith23"  |
| HashedPassword | string |      | "notarealhash" |

### PostgreSQL

| Name                  | Type          | Constraints | Example          | Constraints |
| --------------------- | ------------- | ----------- | ---------------- | ----------- |
| username              | TEXT NOT NULL |             | "JohnSmith23"    | UNIQUE      |
| hashedPassword        | TEXT NOT NULL |             | "notarealhash"   |             |

### JSON

| Name           | Type   | Comments | Example        |
| -------------- | -------| -------- | -------------- |
| username       | string |          | "JohnSmith23"  |
| hashedPassword | string |          | "notarealhash" |
