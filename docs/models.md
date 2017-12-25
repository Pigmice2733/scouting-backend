# Data Types

## BasicEvent

### Go

| Name      | Type          | Tags      | Example                    |
| --------- | ------------- | --------- | -------------------------- |
| Key       | string        |           | "2017cmptx"                |
| Name      | string        |           | "Einstein Field (Houston)" |
| ShortName | string        |           | "Einstein (Houston)"       |
| Date      | time.Time     |           | 2017-07-29T15:20:00Z       |

### PostgreSQL

See [Event](#event)

### JSON

| Name      | Type   | Comments      | Example                    |
| --------- | ------ | ------------- | -------------------------- |
| key       | string |               | "2017cmptx"                |
| name      | string |               | "Einstein Field (Houston)" |
| shortName | string |               | "Einstein (Houston)"       |
| date      | string |               | "2017-07-29T15:20:00Z"     |

---

## Event

### Go

| Name      | Type                            | Tags | Example                         |
| --------- | ------------------------------- | ---- | ------------------------------- |
|           | [event.BasicEvent](#basicevent) |      | See [BasicEvent](#basicevent)   |
| Matches   | [][match.BasicMatch](#basicmatch)         |      | See [Match](#basicmatch)             |

### PostgreSQL

| Name      | Type             | Constraints | Example                    |
| --------- | ---------------- | ----------- | -------------------------- |
| key       | TEXT PRIMARY KEY |             | "2017cmptx"                |
| name      | TEXT             | NOT NULL    | "Einstein Field (Houston)" |
| shortName | TEXT             |             | "Einstein (Houston)"       |
| date      | TIMESTAMPTZ      | NOT NULL    | 2017-07-29T15:20:00Z       |

### JSON

| Name      | Type                    | Comments | Example                    |
| --------- | ----------------------- | -------- | -------------------------- |
| key       | string                  |          | "2017cmptx"                |
| name      | string                  |          | "Einstein Field (Houston)" |
| shortName | string                  |          | "Einstein (Houston)"       |
| date      | int (UNIX TIME)         |          | 1512764281                 |
| matches   | [][match.Basic](#basicmatch) |          | See [Match](#basicmatch)        |

---

## BasicMatch

### Go

| Name              | Type              | Tags      | Example              |
| ----------------- | ----------------- | --------- | -------------------- |
| Key               | string            |           | "2017cmptx_sf1m13"   |
| EventKey          | string            | -         | "2017cmptx"          |
| PredictedTime     | time.Time         | omitempty | 2017-07-29T15:20:00Z |
| ActualTime        | time.Time         | omitempty | 2017-07-29T15:20:00Z |

### PostgreSQL

See [Match](#match)

### JSON

| Name              | Type   | Comments      | Example                |
| ----------------- | ------ | ------------- | ---------------------- |
| key               | string |               | "2017cmptx_sf1m13"     |
| predictedTime     | string | Omit if empty | "2017-07-29T15:20:00Z" |
| actualTime        | string | Omit if empty | "2017-07-29T15:20:00Z" |

---

## Match

### Go

| Name         | Type                            | Tags      | Example                            |
| ------------ | ------------------------------- | --------- | ---------------------------------  |
|              | [match.BasicMatch](#basicmatch) |           | See [BasicMatch](#basicmatch)      |
| BlueWon      | \*bool                          | omitempty | true                               |
| RedAlliance  | []string                        |           | [ "frc1011", "frc5499", "frc973" ] |
| BlueAlliance | []string                        |           | [ "frc1011", "frc5499", "frc973" ] |

### PostgreSQL

| Name            | Type              | Constraints                                  | Example                         |
| --------------- | ----------------- | -------------------------------------------- | ------------------------------- |
| key             | TEXT PRIMARY KEY  |                                              | "2017cmptx_sf1m13"              |
| eventKey        | TEXT NOT NULL     | FOREIGN KEY(eventKey) REFERENCES events(key) | "2017cmptx"                     |
| predictedTime   | TIMESTAMPTZ       |                                              | 2017-04-21 17:00:00 -0700 -0700 |
| actualTime      | TIMESTAMPTZ       |                                              | 2017-04-21 17:00:00 -0700 -0700 |
| blueWon         | BOOLEAN           |                                              | true                            |
| redScore        | INTEGER           |                                              | 83                              |
| blueScore       | INTEGER           |                                              | 96                              |

### JSON

| Name              | Type                           | Comments      | Example                            |
| ----------------- | ------------------------------ | ------------- | ---------------------------------- |
| key               | string                         |               | "2017cmptx_sf1m13"                 |
| predictedTime     | int (UNIX TIME)                | Omit if empty | 1512764281                         |
| actualTime        | int (UNIX TIME)                | Omit if empty | 1512764281                         |
| blueWon           | bool                           | Omit if empty | true                               |
| blueAlliance      | []string                       |               | [ "frc1011", "frc5499", "frc973" ] |
| redAlliance       | []string                       |               | [ "frc1011", "frc5499", "frc973" ] |

---

## Alliance

### Go

An Alliance is just a []string in Go.

### PostgreSQL

| Name           | Type                       | Constraints                                            | Example                         |
| -------------- | -------------------------- | ------------------------------------------------------ | ------------------------------- |
| matchKey       | TEXT NOT NULL              | FOREIGN KEY(matchKey) REFERENCES matches(key)          | "2017cmptx_sf1m13"              |
| isBlue         | BOOLEAN NOT NULL           |                                                        | true                            |
| number         | TEXT NOT NULL              |                                                        | "frc2733b"                      |
|                |                            | UNIQUE(matchKey, number)                               |                                 |

### JSON

An Alliance is just an array of strings (team numbers) in js.

---

## Report

### Go

| Name              | Type                   | Tags | Example            |
| ----------------- | ---------------------- | ---- | ------------------ |
| Reporter          | string                 |      | "JohnSmith2"       |
| EventKey          | string                 |      | "2017cmptx"        |
| MatchKey          | string                 |      | "2017cmptx_sf1m13" |
| IsBlue            | bool                   |      | true               |
| Team              | string                 |      | "frc2740b"         |
| Stats             | map[string]interface{} |      |                    |

### PostgreSQL

| Name              | Type              | Constraints                                               | Example            |
| ----------------- | ----------------- | --------------------------------------------------------- | ------------------ |
| reporter          | TEXT              | NOT NULL FOREIGN KEY(reporter) REFERENCES users(username) | "JohnSmith2"       |
| eventKey          | TEXT              | NOT NULL FOREIGN KEY(eventKey) REFERENCES events(key)     | "2017cmptx"        |
| matchKey          | TEXT              | NOT NULL FOREIGN KEY(matchKey) REFERENCES matches(key)    | "2017cmptx_sf1m13" |
| isBlue            | BOOLEAN           | NOT NULL                                                  | true               |
| team              | TEXT              | NOT NULL                                                  | "frc2740b"         |
| stats             | TEXT              | NOT NULL                                                  |                    |
|                   |                   | UNIQUE(eventKey, matchKey)                                |                    |

### JSON

| Name              | Type   | Comments | Example                                        |
| ----------------- | ------ | -------- | ---------------------------------------------- |
| reporter          | string |          | "JohnSmith2"                                   |
| isBlue            | bool   |          | true                                           |
| team              | string |          | ["frc2471", "frc2733", "frc1418"]              |
| stats             | object |          | { climbed: true, gears: 6, crossedLine: true } |

---

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