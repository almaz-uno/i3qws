#!/bin/bash

go mod tidy

GOOS=linux GOARCH=amd64 go install
