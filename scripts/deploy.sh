#!/bin/bash

ID=$(docker build -q -t scouting-backend ../)
docker tag $ID fharding/scouting-backend
docker push fharding/scouting-backend
