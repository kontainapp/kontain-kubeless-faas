#!/bin/bash -x


rm -f go.mod
rm -f go.sum
rm -f kontain-faas-server

go mod init kontain-faas
GOOS=linux GOARCH=amd64 go build -o kontain-faas-server .
