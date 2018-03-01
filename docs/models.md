# Data Types

## BasicEvent

### Go

| Name      | Type      | Tags      | Example                    |
| --------- | --------- | --------- | -------------------------- |
| Key       | string    |           | "2017cmptx"                |
| Name      | string    |           | "Einstein Field (Houston)" |
| ShortName | string    |           | "Einstein (Houston)"       |
| Date      | time.Time |           | 2017-07-29T15:20:00Z       |
| EndDate   | time.Time |           | 2017-07-29T15:20:00Z       |
| Lat       | \*float64 | omitempty | 42.937225341796875         |
| Long      | \*float64 | omitempty | -71.51953887939453         |
| EventType | int       |           | 99                         |

### PostgreSQL

See [Event](#event)

### JSON

| Name      | Type    | Comments  | Example                    |
| --------- | ------- | --------- | -------------------------- |
| key       | string  |           | "2017cmptx"                |
| name      | string  |           | "Einstein Field (Houston)" |
| shortName | string  |           | "Einstein (Houston)"       |
| date      | string  |           | "2017-07-29T15:20:00Z"     |
| endDate   | string  |           | "2017-07-29T15:20:00Z"     |
| lat       | number  | omitempty | 42.937225341796875         |
| long      | number  | omitempty | -71.51953887939453         |
| eventType | integer |           | 99                         |

---

## Event

### Go

| Name    | Type                              | Tags | Example                       |
| ------- | --------------------------------- | ---- | ----------------------------- |
|         | [event.BasicEvent](#basicevent)   |      | See [BasicEvent](#basicevent) |
| Matches | [][match.basicmatch](#basicmatch) |      | See [Match](#basicmatch)      |

### PostgreSQL

| Name      | Type             | Constraints | Example                    |
| --------- | ---------------- | ----------- | -------------------------- |
| key       | TEXT PRIMARY KEY |             | "2017cmptx"                |
| name      | TEXT             | NOT NULL    | "Einstein Field (Houston)" |
| shortName | TEXT             |             | "Einstein (Houston)"       |
| date      | TIMESTAMPTZ      | NOT NULL    | 2017-07-29T15:20:00Z       |
| endDate   | TIMESTAMPTZ      | NOT NULL    | 2017-07-29T15:20:00Z       |
| Lat       | REAL             |             | 42.937225341796875         |
| Long      | REAL             |             | -71.51953887939453         |
| eventType | INTEGER          | NOT NULL    | 99                         |

### JSON

| Name      | Type                         | Comments  | Example                    |
| --------- | ---------------------------- | --------- | -------------------------- |
| key       | string                       |           | "2017cmptx"                |
| name      | string                       |           | "Einstein Field (Houston)" |
| shortName | string                       |           | "Einstein (Houston)"       |
| date      | int (UNIX TIME)              |           | 1512764281                 |
| endDate   | int (UNIX TIME)              |           | 1512764281                 |
| lat       | number                       | omitempty | 42.937225341796875         |
| long      | number                       | omitempty | -71.51953887939453         |
| eventType | int                          |           | 99                         |
| matches   | [][match.basic](#basicmatch) |           | See [Match](#basicmatch)   |

---

## BasicMatch

### Go

| Name          | Type      | Tags      | Example                                       |
| ------------- | --------- | --------- | --------------------------------------------- |
| Key           | string    |           | "2017cmptx_sf1m13"                            |
| EventKey      | string    | -         | "2017cmptx"                                   |
| PredictedTime | time.Time | omitempty | 2017-07-29T15:20:00Z                          |
| ActualTime    | time.Time | omitempty | 2017-07-29T15:20:00Z                          |
| YoutubeURL    | string    |           | "https://www.youtube.com/watch?v=dTjzn4HCP-o" |

### PostgreSQL

See [Match](#match)

### JSON

| Name          | Type   | Comments      | Example                                       |
| ------------- | ------ | ------------- | --------------------------------------------- |
| key           | string |               | "2017cmptx_sf1m13"                            |
| predictedTime | string | Omit if empty | "2017-07-29T15:20:00Z"                        |
| actualTime    | string | Omit if empty | "2017-07-29T15:20:00Z"                        |
| youtubeURL    | string |               | "https://www.youtube.com/watch?v=dTjzn4HCP-o" |

---

## Match

### Go

| Name         | Type                            | Tags | Example                            |
| ------------ | ------------------------------- | ---- | ---------------------------------- |
|              | [match.BasicMatch](#basicmatch) |      | See [BasicMatch](#basicmatch)      |
| RedAlliance  | []string                        |      | [ "frc1011", "frc5499", "frc973" ] |
| BlueAlliance | []string                        |      | [ "frc1011", "frc5499", "frc973" ] |

### PostgreSQL

| Name          | Type             | Constraints                                  | Example                                       |
| ------------- | ---------------- | -------------------------------------------- | --------------------------------------------- |
| key           | TEXT PRIMARY KEY |                                              | "2017cmptx_sf1m13"                            |
| eventKey      | TEXT NOT NULL    | FOREIGN KEY(eventKey) REFERENCES events(key) | "2017cmptx"                                   |
| predictedTime | TIMESTAMPTZ      |                                              | 2017-04-21 17:00:00 -0700 -0700               |
| actualTime    | TIMESTAMPTZ      |                                              | 2017-04-21 17:00:00 -0700 -0700               |
| redScore      | INTEGER          |                                              | 83                                            |
| blueScore     | INTEGER          |                                              | 96                                            |
| youtubeURL    | TEXT             |                                              | "https://www.youtube.com/watch?v=dTjzn4HCP-o" |

### JSON

| Name          | Type            | Comments      | Example                            |
| ------------- | --------------- | ------------- | ---------------------------------- |
| key           | string          |               | "2017cmptx_sf1m13"                 |
| predictedTime | int (UNIX TIME) | Omit if empty | 1512764281                         |
| actualTime    | int (UNIX TIME) | Omit if empty | 1512764281                         |
| blueAlliance  | []string        |               | [ "frc1011", "frc5499", "frc973" ] |
| redAlliance   | []string        |               | [ "frc1011", "frc5499", "frc973" ] |

---

## Alliance

### Go

An Alliance is just a []string in Go.

### PostgreSQL

| Name     | Type             | Constraints                                   | Example            |
| -------- | ---------------- | --------------------------------------------- | ------------------ |
| matchKey | TEXT NOT NULL    | FOREIGN KEY(matchKey) REFERENCES matches(key) | "2017cmptx_sf1m13" |
| isBlue   | BOOLEAN NOT NULL |                                               | true               |
| number   | TEXT NOT NULL    |                                               | "frc2733b"         |
|          |                  | UNIQUE(matchKey, number)                      |                    |

### JSON

An Alliance is just an array of strings (team numbers) in js.

---

## Report

### Go

| Name     | Type                   | Tags | Example            |
| -------- | ---------------------- | ---- | ------------------ |
| Reporter | string                 |      | "JohnSmith2"       |
| EventKey | string                 |      | "2017cmptx"        |
| MatchKey | string                 |      | "2017cmptx_sf1m13" |
| IsBlue   | bool                   |      | true               |
| Team     | string                 |      | "frc2740b"         |
| Stats    | map[string]interface{} |      |                    |

### PostgreSQL

| Name     | Type    | Constraints                                               | Example            |
| -------- | ------- | --------------------------------------------------------- | ------------------ |
| reporter | TEXT    | NOT NULL FOREIGN KEY(reporter) REFERENCES users(username) | "JohnSmith2"       |
| eventKey | TEXT    | NOT NULL FOREIGN KEY(eventKey) REFERENCES events(key)     | "2017cmptx"        |
| matchKey | TEXT    | NOT NULL FOREIGN KEY(matchKey) REFERENCES matches(key)    | "2017cmptx_sf1m13" |
| isBlue   | BOOLEAN | NOT NULL                                                  | true               |
| team     | TEXT    | NOT NULL                                                  | "frc2740b"         |
| stats    | TEXT    | NOT NULL                                                  |                    |
|          |         | UNIQUE(eventKey, matchKey)                                |                    |

### JSON

| Name     | Type   | Comments | Example                                        |
| -------- | ------ | -------- | ---------------------------------------------- |
| reporter | string |          | "JohnSmith2"                                   |
| isBlue   | bool   |          | true                                           |
| team     | string |          | ["frc2471", "frc2733", "frc1418"]              |
| stats    | object |          | { climbed: true, gears: 6, crossedLine: true } |

---

## User

### Go

| Name           | Type   | Tags | Example        |
| -------------- | ------ | ---- | -------------- |
| Username       | string |      | "JohnSmith23"  |
| HashedPassword | string |      | "notarealhash" |
| IsAdmin        | bool   |      | true           |

### PostgreSQL

| Name           | Type          | Constraints | Example        |
| -------------- | ------------- | ----------- | -------------- |
| username       | TEXT NOT NULL | UNIQUE      | "JohnSmith23"  |
| hashedPassword | TEXT NOT NULL |             | "notarealhash" |
| isAdmin        | TEXT NOT NULL |             | true           |

### JSON

| Name           | Type   | Comments | Example        |
| -------------- | ------ | -------- | -------------- |
| username       | string |          | "JohnSmith23"  |
| hashedPassword | string |          | "notarealhash" |
| isAdmin        | bool   |          | true           |

--

## Photo

### PostgreSQL

| Name | Type          | Constraints       | Example                           |
| ---- | ------------- | ----------------- | --------------------------------- |
| team | TEXT NOT NULL |                   | "frc1678"                         |
| url  | TEXT NOT NULL |                   | "http://i.imgur.com/uN3ojZyl.jpg" |
|      |               | UNIQUE(team, url) |                                   |

--

## JWT Body

### JSON

| Name             | Type   | Example    |
| ---------------- | ------ | ---------- |
| sub              | string | "test"     |
| exp              | int    | 1516739340 |
| pigmice_is_admin | bool   | true       |
