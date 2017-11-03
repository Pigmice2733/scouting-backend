# Pigmice Scouting Backend

## Installing
Clone from Github

Install [Go](https://golang.org/doc/install)

Create `config.json` in the project directory (*go/src/scouting-backend*), containing
```json
{
    "tbaApikey": "<your key>"
}
```
To get an apikey for TBA, go to [thebluealliance/mytba](https://www.thebluealliance.com/mytba) and request an API key for the READ API v3. This is your personal API key, and **must not** be shared online or put in source control. **Keep it personal and secret.**


## Running
To build, run `go build` in the project directory.
To run tests, use `go test` or `go test -v` for verbose logging.
To run the backend run the executable generated by `go build`.
