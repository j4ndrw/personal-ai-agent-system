#!/bin/bash

cd client
go build ./cmd/main.go
mkdir -p .build
mv main .build/paias
cd ..
