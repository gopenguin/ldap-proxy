#!/bin/bash

CGO_ENABLED=0 GOOS=linux go build -o ldap-proxy-static -a -ldflags '-extldflags "-static"' .
