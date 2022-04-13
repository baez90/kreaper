#!/usr/bin/env bash

CGO_ENABLED=0 go build -trimpath -ldflags "-w -s" -installsuffix cgo -o dist/kreaper main.go