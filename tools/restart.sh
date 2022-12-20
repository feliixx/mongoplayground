#!/bin/bash

kill "$(lsof -t -i:8080 -sTCP:LISTEN)"

( cd internal/web && ./bundle.sh )

go build && ./mongoplayground&
