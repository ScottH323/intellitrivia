#!/bin/sh

echo Building Service
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .
if [ $? -eq 0 ]; then
    echo "\033[32mBUILD OK \033[0m"
else
    echo "\033[31mERROR: BUILD FAILED \033[0m"
    exit 1
fi
