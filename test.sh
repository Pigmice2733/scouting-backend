#!/bin/bash

docker-compose -f compose-testing.yml down
docker-compose -f compose-testing.yml build
docker-compose -f compose-testing.yml up -d postgres
docker-compose -f compose-testing.yml run go
