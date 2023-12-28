#!/bin/env sh
cd exporter
go run exporter.go
cp -f font.go ../
cd ..
go build -o tablet-mapper .
