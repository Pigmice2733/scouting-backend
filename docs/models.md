# Data Types

## BasicEvent

### Go

| Name      | Type          | Tags      | Example                         |
| --------- | ------------- | --------- | ------------------------------- |
| Key       | string        |           | "2017cmptx"                     |
| Name      | string        |           | "Einstein Field (Houston)"      |
| ShortName | string        |           | "Einstein (Houston)"            |
| Date      | time.Time     |           | 2017-04-21 17:00:00 -0700 -0700 |

### PostgreSQL

See [Event](#event)

### JSON

| Name      | Type            | Comments      | Example                    |
| --------- | ----------------| ------------- | -------------------------- |
| key       | string          |               | "2017cmptx"                |
| name      | string          |               | "Einstein Field (Houston)" |
| shortName | string          |               | "Einstein (Houston)"       |
| date      | int (UNIX TIME) |               | 1512764281                 |


## Event

### Go

| Name      | Type                            | Tags | Example                         |
| --------- | ------------------------------- | ---- | ------------------------------- |
|           | [event.BasicEvent](#basicevent) |      | See [BasicEvent](#basicevent)   |
| Matches   | [][match.Match](#match)         |      | See [Match](#match)             |

### PostgreSQL

| Name      | Type             | Constraints | Example                         |
| --------- | ---------------- | ----------- | ------------------------------- |
| key       | TEXT PRIMARY KEY |             | "2017cmptx"                     |
| name      | TEXT             | NOT NULL    | "Einstein Field (Houston)"      |
| shortName | TEXT             |             | "Einstein (Houston)"            |
| date      | TIMESTAMPTZ      | NOT NULL    | 2017-04-21 17:00:00 -0700 -0700 |

### JSON

| Name      | Type                    | Comments | Example                    |
| --------- | ----------------------- | -------- | -------------------------- |
| key       | string                  |          | "2017cmptx"                |
| name      | string                  |          | "Einstein Field (Houston)" |
| shortName | string                  |          | "Einstein (Houston)"       |
| date      | int (UNIX TIME)         |          | 1512764281                 |
| matches   | [][match.Match](#match) |          | See [Match](#match)        |


## BasicMatch

### Go

| Name              | Type              | Tags      | Example                           |
| ----------------- | ----------------- | --------- | --------------------------------- |
| Key               | string            |           | "2017cmptx_sf1m13"                |
| EventKey          | string            | -         | "2017cmptx"                       |
| PredictedTime     | time.Time         | omitempty | 2017-04-21 17:00:00 -0700 -0700   |
| ActualTime        | time.Time         | omitempty | 2017-04-21 17:00:00 -0700 -0700   |

### PostgreSQL

See [Match](#match)

### JSON

| Name              | Type              | Comments      | Example                           |
| ----------------- | ----------------- | ------------- | --------------------------------- |
| key               | string            |               | "2017cmptx_sf1m13"                |
| predictedTime     | int (UNIX TIME)   | Omit if empty | 1512764281                        |
| actualTime        | int (UNIX TIME)   | Omit if empty | 1512764281                        |


## Match

### Go

| Name         | Type                            | Tags      | Example                           |
| ------------ | ------------------------------- | --------- | --------------------------------- |
|              | [match.BasicMatch](#basicmatch) |           | See [BasicMatch](#basicmatch)     |
| BlueWon      | \*bool                          | omitempty | true                              |
| RedAlliance  | [alliance.Alliance](#alliance)  |           | See [Alliance](#alliance)         |
| BlueAlliance | [alliance.Alliance](#alliance)  |           | See [Alliance](#alliance)         |

### PostgreSQL

| Name            | Type              | Constraints                                  | Example                         |
| --------------- | ----------------- | -------------------------------------------- | ------------------------------- |
| key             | TEXT PRIMARY KEY  |                                              | "2017cmptx_sf1m13"              |
| eventKey        | TEXT NOT NULL     | FOREIGN KEY(eventKey) REFERENCES events(key) | "2017cmptx"                     |
| predictedTime   | TIMESTAMPTZ       |                                              | 2017-04-21 17:00:00 -0700 -0700 |
| actualTime      | TIMESTAMPTZ       |                                              | 2017-04-21 17:00:00 -0700 -0700 |
| winningAlliance | TEXT              |                                              | "blue"                          |
| redScore        | INTEGER           |                                              | 83                              |
| blueScore       | INTEGER           |                                              | 96                              |

### JSON

| Name              | Type                           | Comments      | Example                   |
| ----------------- | ------------------------------ | ------------- | ------------------------- |
| key               | string                         |               | "2017cmptx_sf1m13"        |
| predictedTime     | int (UNIX TIME)                | Omit if empty | 1512764281                |
| actualTime        | int (UNIX TIME)                | Omit if empty | 1512764281                |
| blueWon           | bool                           | Omit if empty | true                      |
| redAlliance       | [alliance.Alliance](#alliance) |               | See [Alliance](#alliance) |
| blueAlliance      | [alliance.Alliance](#alliance) |               | See [Alliance](#alliance) |


## Alliance

| Name              | Type                   | Tags | Example                           |
| ----------------- | ---------------------- | ---- | --------------------------------- |
| Score             | int                    |      | 96                                |
| IsBlue            | bool                   |      | true                              |
| Teams             | []string               |      | ["frc2471", "frc2733", "frc1418"] |

### PostgreSQL

Not stored.

## JSON

| Name              | Type     | Comments | Example                           |
| ----------------- | -------- | -------- | --------------------------------- |
| score             | int      |          | 96                                |
| isBlue            | bool     |          | true                              |
| teams             | []string |          | ["frc2471", "frc2733", "frc1418"] |

## Team

### PostgreSQL

| Name           | Type                       | Constraints                                            | Example                         |
| -------------- | -------------------------- | ------------------------------------------------------ | ------------------------------- |
| matchKey       | TEXT                       | NOT NULL FOREIGN KEY(matchKey) REFERENCES matches(key) | "2017cmptx_sf1m13"              |
| isAllianceBlue | BOOLEAN                    | NOT NULL                                               | true                            |
| number         | TEXT                       | NOT NULL                                               | "frc2733b"                      |


## Report

### Go

| Name              | Type                   | Tags | Example            |
| ----------------- | ---------------------- | ---- | ------------------ |
| Reporter          | string                 |      | "JohnSmith2"       |
| EventKey          | string                 |      | "2017cmptx"        |
| MatchKey          | string                 |      | "2017cmptx_sf1m13" |
| IsAllianceBlue    | bool                   |      | true               |
| Team              | string                 |      | "frc2740b"         |
| Stats             | map[string]interface{} |      |                    |

### PostgreSQL

| Name              | Type              | Constraints                                               | Example            |
| ----------------- | ----------------- | --------------------------------------------------------- | ------------------ |
| reporter          | TEXT              | NOT NULL FOREIGN KEY(reporter) REFERENCES users(username) | "JohnSmith2"       |
| eventKey          | TEXT              | NOT NULL FOREIGN KEY(eventKey) REFERENCES events(key)     | "2017cmptx"        |
| matchKey          | TEXT              | NOT NULL FOREIGN KEY(matchKey) REFERENCES matches(key)    | "2017cmptx_sf1m13" |
| isAllianceBlue    | BOOLEAN           | NOT NULL                                                  | true               |
| team              | TEXT              | NOT NULL                                                  | "frc2740b"         |
| stats             | TEXT              | NOT NULL                                                  |                    |
|                   |                   | UNIQUE(eventKey, matchKey)                                |                    |

### JSON

| Name              | Type   | Comments | Example                                        |
| ----------------- | ------ | -------- | ---------------------------------------------- |
| reporter          | string |          | "JohnSmith2"                                   |
| isAllianceBlue    | bool   |          | true                                           |
| team              | string |          | ["frc2471", "frc2733", "frc1418"]              |
| stats             | object |          | { climbed: true, gears: 6, crossedLine: true } |


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


## User

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
