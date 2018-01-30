#!/bin/bash

export PG_SSL_MODE=disable
export PG_USER=scoutingbackend
export PG_DB_NAME=scoutingbackend
export HTTP_ADDR=:8080

cd ../cmd/scouting-backend
go build
./scouting-backend
