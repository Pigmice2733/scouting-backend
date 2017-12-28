#!/bin/bash

cd ../internal/store/postgres/migrations
go-bindata -pkg migrations .
