#!/usr/bin/env bash

GOOS=linux GOARCH=amd64 go build -o bin/builder ./cmd/builder
