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

## /users - POST

Creates a new user. Usernames can only contain alphanumeric + space characters. Users created by admins will be created instantly. Accounts created by unauthenticated users are created but require verification by an admin before they can be used.

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
      "actualTime": "2018-02-17T12:58:26-08:00",
      "youtubeURL": "https://www.youtube.com/watch?v=dTjzn4HCP-o"
    },
    {
      "key": "2018week0_f1m2",
      "predictedTime": "2018-02-17T13:11:01-08:00",
      "actualTime": "2018-02-17T13:11:59-08:00",
      "youtubeURL": "https://www.youtube.com/watch?v=dTjzn4HCP-o"
    },
    ...
  ]
}
```

---

## /events/{eventKey}/matches/{matchKey} - GET

Gets a complete match.

### Response Body

```json
{
  "key": "2017nhfoc_f1m1",
  "predictedTime": "2017-07-29T15:20:00Z",
  "actualTime": "2017-07-29T15:20:42Z",
  "youtubeURL": "https://www.youtube.com/watch?v=dTjzn4HCP-o",
  "redScore": 508,
  "blueScore": 342,
  "redAlliance": ["frc1011", "frc5499", "frc973"],
  "blueAlliance": ["frc1011", "frc5499", "frc973"]
}
```

---

## /events/{eventKey}/matches/{matchKey}/reports - PUT - Authenticated

Upserts a report

The request body can change depending on the schema and data to analyze for the stats field.

### Request Body

```json
{
  "team": "frc2733",
  "notes": "notes on the team",
  "stats": {
    "climbed": true,
    "movedBunnies": 10,
    "movedBuckets": 5
  }
}
```

---

## /events/{eventKey}/teams/{team}/reports - GET

Retrieve all reports for the specified team and event

### Response Body

```json
[
  {
    "reporter": "JohnSmith2",
    "team": "frc2733",
    "notes": "notes on the team",
    "stats": {
      "climbed": true,
      "movedBunnies": 10,
      "movedBuckets": 5
    }
  }
]
```

---

## /teams/{team}/reports - GET

Retrieve all reports for a team.

```json
[
  {
    "reporter": "Dexter",
    "eventKey": "2018orwil",
    "matchKey": "2018orwil_qm3",
    "team": "frc4488",
    "notes": null,
    "stats": {
      "autoCrossedLine": true,
      "autoCubesOnScale": 0,
      "autoCubesOnSwitch": 0,
      "hadConnectionProblems": true,
      "hadPowerProblems": false,
      "present": true,
      "teleopClimbed": false,
      "teleopCubesIntoExchange": 1,
      "teleopCubesOnScale": 1,
      "teleopCubesOnSwitch": 1,
      "teleopEndsOnPlatform": false
    }
  }
  ...
]
```

---

## /events/{eventKey}/analysis - GET

Stats about how all teams in an event have performed on average.

### Response Body

The response body can change depending on the schema and data to analyze.

```json
[
  {
    "team": "frc4905",
    "notes": { "2018week0_qm10": "asdf" },
    "reports": 6,
    "stats": {
      "autoCrossedLine": 0,
      "autoCubesOnScale": 0,
      "autoCubesOnSwitch": 0,
      "hadConnectionProblems": 0,
      "hadPowerProblems": 0,
      ...
    }
  }
]
```

---

## /events/{eventKey}/teams/{team}/analysis - GET

Stats about how a team has performed at an event on average.

### Response Body

The response body can change depending on the schema and data to analyze.

```json
{
  "team": "frc4905",
  "notes": { "2018week0_qm10": "asdf" },
  "reports": 6,
  "stats": {
    "autoCrossedLine": 0,
    "autoCubesOnScale": 0,
    "autoCubesOnSwitch": 0,
    "hadConnectionProblems": 0,
    "hadPowerProblems": 0,
    ...
  }
}
```

---

## /events/{eventKey}/matches/{matchKey}/alliance/{color}/analysis - GET

Stats about how all teams on an alliance have performed at an event on average.

### Response Body

The response body can change depending on the schema and data to analyze.

```json
[
  {
    "team": "frc4905",
    "notes": { "2018week0_qm10": "asdf" },
    "reports": 6,
    "stats": {
      "autoCrossedLine": 0,
      "autoCubesOnScale": 0,
      "autoCubesOnSwitch": 0,
      "hadConnectionProblems": 0,
      "hadPowerProblems": 0,
      ...
    }
  }
]
```

---

## /picklists - GET - Authenticated

Retrieves all of the authenticated users basic picklist info.

### Response Body

```json
[
  {
    "id": "de3f0e41-8fde-45d5-b7b4-cb5e06c0584d",
    "eventKey": "2018orwil",
    "name": "climbers",
  },
  {
    "id": "de3f0e41-8fde-45d5-b7b4-cb5e06c0584e",
    "eventKey": "2018orwil",
    "name": "switch"
  },
  {
    "id": "de3f0e41-8fde-45d5-b7b4-cb5e06c0584f",
    "eventKey": "2018orore",
    "name": "switch"
  }
  ...
]
```

---

## /picklists/event/{eventKey} - GET - Authenticated

Retrieves all of the authenticated users basic picklist info for a specific event.

### Response Body

```json
[
  {
    "id": "de3f0e41-8fde-45d5-b7b4-cb5e06c0584a",
    "eventKey": "2018orwil",
    "name": "climbers"
  },
  {
    "id": "de3f0e41-8fde-45d5-b7b4-cb5e06c0584o",
    "eventKey": "2018orwil",
    "name": "switch"
  }
  ...
]
```

---

## /picklists - POST - Authenticated

Creates a new picklist for the authenticated user.

### Request Body

```json
{
  "eventKey": "2018orwil",
  "name": "climbers",
  "list": ["frc2733", "frc2471", "frc254"]
}
```

### Response Body

```json
"de3f0e41-8fde-45d5-b7b4-cb5e06c0584d"
```

---

## /picklists/{id} - GET

Gets a picklist with a given ID.

### Response Body

```json
{
  "id": "de3f0e41-8fde-45d5-b7b4-cb5e06c0584d",
  "eventKey": "2018orore",
  "name": "switch",
  "list": ["frc2733", "frc2471", "frc118"],
  "owner": "franklin"
}
```

---

## /picklists/{id} - PUT - Authenticated (and resource belongs to authenticated user)

Updates a picklist with a given ID.

### Request Body

```json
{
  "eventKey": "2018orore",
  "name": "switch",
  "list": ["frc2733", "frc2471", "frc254"]
}
```

---

## /picklists/{id} - DELETE - Authenticated (and resource belongs to the authenticated user)

Deletes a picklist with a given ID.

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

## /events/{eventKey}/teams - GET

Gets all teams at an event that have been reported on.

## Response Body

```json
["frc2733", "frc2471", "frc254"]
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
