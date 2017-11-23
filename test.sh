#!/bin/bash

docker-compose -f compose-testing.yml down
docker-compose -f compose-testing.yml build
docker-compose -f compose-testing.yml up -d postgres
sleep 2
docker-compose -f compose-testing.yml run go
