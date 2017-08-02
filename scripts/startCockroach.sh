#!/bin/bash

mkdir -p cockroach-data/roach1

docker run -d \
--name=roach1 \
--hostname=roach1 \
-p 26257:26257 -p 8080:8080  \
-v "${PWD}/cockroach-data/roach1:/cockroach/cockroach-data"  \
cockroachdb/cockroach:v1.0.3 start --insecure

sleep 10

psql --host localhost --port 26257 -U root -w -c "CREATE USER test WITH PASSWORD 'test'; CREATE DATABASE auth; GRANT ALL ON DATABASE auth TO test;"

