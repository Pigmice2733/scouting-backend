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
  "jwt":
    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTMyMTA0OTksInN1YiI6ImZyYW5rIn0.HCwmwj0f-4b2I-hK9QEJ-4berztETH_IDgcSIJBPXMI"
}
```

---

## /users - GET - Authenticated (Admin Users Only)

Gets all users.

### Response Body

```json
[
  { "isAdmin": false, "username": "test2" },
  { "isAdmin": true, "username": "test" }
]
```

---

## /users - POST - Authenticated (Admin Users Only)

Creates a new user.

### Request Body

```json
{
  "username": "frank2",
  "password": "asdf",
  "isAdmin": false
}
```

---

## /users/{username} - DELETE - Authenticated (Admin Users Only)

Deletes a user.

---

## /users/{username} - PUT - Authenticated

Updates a user.

### Request Body

```
{
  "username": "asdf",
  "password": "test",
  "isAdmin": false
}
```

---

## /events - GET

Gets all (basic) events.

### Response Body

```json
[
  {
    "key": "2018week0",
    "name": "Week 0",
    "shortName": "Week 0",
    "eventType": 100,
    "lat": 42.937225341796875,
    "long": -71.51953887939453,
    "date": "2018-02-16T16:00:00-08:00",
    "endDate": "2018-02-18T16:00:00-08:00"
  },
  {
    "key": "2018wila",
    "name": "Seven Rivers Regional",
    "shortName": "Seven Rivers",
    "eventType": 0,
    "lat": 43.812232971191406,
    "long": -91.25572204589844,
    "date": "2018-04-03T17:00:00-07:00",
    "endDate": "2018-04-15T16:00:00-08:00"
  },
  ...
]
```

---

## /events/{eventKey} - GET

Gets a complete event including matches.

```json
{
  "key": "2018week0",
  "name": "Week 0",
  "shortName": "Week 0",
  "eventType": 100,
  "lat": 42.937225341796875,
  "long": -71.51953887939453,
  "date": "2018-02-16T16:00:00-08:00",
  "endDate": "2018-02-17T16:00:00-08:00",
  "matches": [
    {
      "key": "2018week0_f1m1",
      "predictedTime": "2018-02-17T12:57:09-08:00",
      "actualTime": "2018-02-17T12:58:26-08:00"
    },
    {
      "key": "2018week0_f1m2",
      "predictedTime": "2018-02-17T13:11:01-08:00",
      "actualTime": "2018-02-17T13:11:59-08:00"
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
  "redScore": 508,
  "blueScore": 342,
  "redAlliance": ["frc1011", "frc5499", "frc973"],
  "blueAlliance": ["frc1011", "frc5499", "frc973"]
}
```

---

## /reports/{eventKey}/{matchKey} - PUT - Authenticated

Upserts a report

The request body can change depending on the schema and data to analyze for the stats field.

### Request Body

```json
{
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
  }
]
```

---

## /analysis/{eventKey}/{team} - GET

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

## /analysis/{eventKey}/{matchKey}/{color} - GET

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

---

## /schema

Sends the report schema.

### Response Body

```json
{
  "climbed": "bool",
  "movedBunnies": "number",
  "movedBuckets": "number"
}
```

---

## /photo/{team}

Responds with the (binary) photo for given team.

### Response Body

`binary photo`

---

## /leaderboard

Responds with the leaderboard of top reporters.

### Response Body

```json
[{ "reporter": "test", "reports": 2 }, { "reporter": "test2", "reports": 4 }]
```
