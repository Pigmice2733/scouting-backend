# Documentation for Scouting Backend HTTP API

## Authenticated Requests

Authenticated requests should have a Authentication header with the format "Authentication: Bearer {signed jwt string}".

Almost all request that have to deal with users or are not a GET request are authenticated with few exceptions.

---

## /authenticate - POST

For retrieving a JWT token for authenticated requests.

### Request Body

```json
{
  "username": "frank",
  "password": "asdf"
}
```

### Response Body

```json
{
  "jwt": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTMyMTA0OTksInN1YiI6ImZyYW5rIn0.HCwmwj0f-4b2I-hK9QEJ-4berztETH_IDgcSIJBPXMI"
}
```

---

## /users - GET - Authenticated

Gets all users.

### Response Body

```json
[
  {
    "username": "frank",
    "hashedPassword": "$2b$12$29H98PpF9E/CccCc.OEBfObX4Sm0coLZE6cCWSAmaNQkveGhop5lW"
  },
  ...
]
```

---

## /users - POST - Authenticated

Creates a new user.

### Request Body

```json
{
  "username": "frank2",
  "password": "asdf"
}
```

---

## /users/{username} - GET - Authenticated

Gets a user.

### Response Body

```json
{
  "username": "frank",
  "hashedPassword": "$2b$12$29H98PpF9E/CccCc.OEBfObX4Sm0coLZE6cCWSAmaNQkveGhop5lW"
}
```

---

## /users/{username} - DELETE - Authenticated

Deletes a user.

---

## /events - GET

Gets all (basic) events.

### Response Body

```json
[
  {
    "key": "2017alhu",
    "name": "Rocket City Regional",
    "shortName": "Rocket City",
    "date": "2017-03-21T17:00:00-07:00"
  },
  {
    "key": "2017code",
    "name": "Colorado Regional",
    "shortName": "Colorado",
    "date": "2017-03-21T17:00:00-07:00"
  },
  ...
]
```

---

## /events/{eventKey} - GET

Gets a complete event including matches.

```json
{
  "key": "2017nhfoc",
  "name": "FIRST Festival of Champions",
  "shortName": "FIRST Festival of Champions",
  "date": "2017-07-28T17:00:00-07:00",
  "matches": [
    {
      "key": "2017nhfoc_f1m1",
      "predictedTime": "2017-07-29T15:20:00Z",
      "actualTime": "2017-07-29T15:20:42Z",
    },
    ...
  ]
}
```

---

## /events/{eventKey}/{matchKey} - GET

Gets a complete match.

### Response Body

```json
{
  "key": "2017nhfoc_f1m1",
  "predictedTime": "2017-07-29T15:20:00Z",
  "actualTime": "2017-07-29T15:20:42Z",
  "blueWon": true,
  "redAlliance": {
    "score": 508,
    "teams": [ "frc1011", "frc5499", "frc973" ]
  },
  "blueAlliance": {
    "score": 342,
    "teams": [ "frc254", "frc2767", "frc1676" ]
  }
}
```

---

## /reports/{eventKey}/{matchKey} - PUT - Authenticated

Upserts a report

### Request Body

```json
{
  "reporter": "frank",
  "isBlueAlliance": true,
  "team": "frc2733",
  "stats": {
    "climbed": true,
    "movedBunnies": 10,
    "movedBuckets": 5
  }
}
```
---

## /analysis/{eventKey} - GET

Stats about how all teams in an event have performed on average.

### Response Body

The response body can change depending on the schema and data to analyze.

```json
[
  {
    "team": "frc254",
    "stats": {
      "climbed": 0.94,
      "movedBunnies": 7.4,
      "movedBuckets": 14.2
    }
  },
  {
    "team": "frc2733",
    "stats": {
      "climbed": 1,
      "movedBunnies": 68.6,
      "movedBuckets": 52.3
    }
  },
  {
    "team": "frc2471",
    "stats": {
      "climbed": 1,
      "movedBunnies": 67.6,
      "movedBuckets": 51.3
    }
  },
  ...
]
```

---

## /analysis/{eventKey}/{teamNumber} - GET

Stats about how a team has performed at an event on average.

### Response Body

The response body can change depending on the schema and data to analyze.

```json
{
  "climbed": 0.94,
  "movedBunnies": 7.4,
  "movedBuckets": 14.2
}
```

---

## /analysis/{eventKey}/{matchKey}/{allianceColor} - GET

Stats about how all teams on an alliance have performed at an event on average.

### Response Body

The response body can change depending on the schema and data to analyze.

```json
[
  {
    "team": "frc254",
    "stats": {
      "climbed": 0.94,
      "movedBunnies": 7.4,
      "movedBuckets": 14.2
    }
  },
  {
    "team": "frc2733",
    "stats": {
      "climbed": 1,
      "movedBunnies": 68.6,
      "movedBuckets": 52.3
    }
  },
  {
    "team": "frc2471",
    "stats": {
      "climbed": 1,
      "movedBunnies": 67.6,
      "movedBuckets": 51.3
    }
  }
]
```
