# Pigmice (FRC 2733) Scouting Backend [![Go Report Card](https://goreportcard.com/badge/github.com/pigmice2733/scouting-backend)](https://goreportcard.com/report/github.com/pigmice2733/scouting-backend) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Environment Variables

Some environment variables are needed to tell the app what to do.

* PG_USER: postgres user
* PG_PASS: postgres password
* PG_HOST: postgres host address
* PG_PORT: postgres port (defaults to 5432)
* PG_DB_NAME: postgres database name
* PG_SSL_MODE: postgres ssl mode
* TBA_API_KEY: the blue alliance api key
* SCHEMA_PATH: path to the report schema
* HTTP_ADDR: http address
* HTTPS_ADDR: https address
* CERT_FILE: path to ssl certificate file
* KEY_FILE: path to ssl key file
* ORIGIN: ACCESS-CONTROL-ALLOW-ORIGIN http header value (defaults to '\*')
* YEAR: year to use when consuming data from tba api (defaults to current year)

## Running

* Go to main directory for the server: `cd cmd/scouting-backend`
* Build: `go build`
* Run: `./scouting-backend`

## Pushing to Docker Hub

* Build the docker image: `docker build -t scouting-backend .`
* Get the docker image ID (the one we just created): `docker images`
* Tag the docker image: `docker tag {docker id} fharding/scouting-backend:latest`
* Push the docker image: `docker push fharding/scouting-backend:latest`

Keep in mind you will obviously need access to fharding/scouting-backend.
Docker Cloud should automatically re-deploy.
