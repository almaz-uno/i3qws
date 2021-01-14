#!/bin/bash

GOOS=linux GOARCH=amd64 go install
#GOOS=darwin GOARCH=amd64 go build -o ${binary}.darwin
#GOOS=windows GOARCH=amd64 go build -o ${binary}.windows.exe
