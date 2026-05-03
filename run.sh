#!/usr/bin/env bash

# exit immediately on any failed command
set -e

# load the env file: -a makes every assignment auto-exported
echo "start .env load"
set -a
source .env
set +a
echo "database in use: $SQLITE_DB_FILENAME"

# generate frontend files with templ 
templ generate

# backend test
go clean -testcache
go test ./data/dbscripts/initDB
go test ./back/api/test/users
go test ./back/api/test/params
go test ./back/api/test/records
go test ./back/api/test/save
go test ./back/api/test/shutdown
go test ./back/api/test/csv

# run server
echo "run server with: $SQLITE_DB_FILENAME"
go run .

# now exec frontend tests with playwright
# README in e2e folder