#!/bin/bash

docker-compose -f $1 down
docker-compose -f $1 build
docker-compose -f $1 up -d postgres
sleep 3
docker-compose -f $1 run go
docker-compose -f $1 down
