#!/bin/bash
( cd internal/web && ./bundle.sh )
go build
kill "$(lsof -t -i:8080 -sTCP:LISTEN)"
 ./mongoplayground&