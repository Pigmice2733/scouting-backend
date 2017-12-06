# Pigmice (FRC 2733) Scouting Backend

## Environment Variables

Some environment variables are needed to tell the app what to connect to.

- PG_USER: postgres user
- PG_PASS: postgres password
- PG_HOST: postgres host address
- PG_PORT: postgres port
- PG_DB_NAME: postgres database name
- PG_SSL_MODE: postgres ssl mode
- PG_MAX_CONNECTIONS: postgres maximum connections
- TBA_API_KEY: the blue alliance api key
- HTTP_ADDR: http address
- HTTPS_ADDR: https address
- CERT_FILE: path to ssl certificate file
- KEY_FILE: path to ssl key file

## Running

- Go to main directory for the server: `cd cmd/scouting-backend`
- Build: `go build`
- Run: `./scouting-backend`
